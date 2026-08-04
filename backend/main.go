package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"tracker/bootstrap"
	pricev1 "tracker/gen/price/v1"
	walletv1 "tracker/gen/wallet/v1"
	"tracker/internal/api/handlers"
	"tracker/internal/api/middleware"
	"tracker/internal/cache"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/bitcoin"
	"tracker/internal/client/coingecko"
	"tracker/internal/client/ethereum"
	"tracker/internal/client/solana"
	"tracker/internal/client/tron"
	"tracker/internal/core"
	"tracker/internal/db"
	"tracker/internal/db/repositories"
	"tracker/internal/docs"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	app := bootstrap.App()

	db, err := db.NewDatabase(app.Cfg.DSN())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Postgres")
	}
	defer db.Close()

	redisClient := cache.NewRedisClient(
		app.Cfg.Redis.Addr,
		app.Cfg.Redis.Password,
		app.Cfg.Redis.DB,
	)
	priceRepo := repositories.NewPriceRepository(db)
	walletRepo := repositories.NewWalletRepository(db)
	userRepo := repositories.NewUserRepo(db)

	coingeckoClient := coingecko.NewCoinGeckoClient(app.Cfg.CoinGecko.APIKey)
	alchemyClient := alchemy.NewAlchemyClient(app.Cfg.Alchemy.APIKey)

	ethMainnetClient := ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.EthereumMainnet, Decimal: 18})
	ethArbitrumClient := ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.EthereumArbitrum, Decimal: 18})
	ethBaseClient := ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.EthereumBase, Decimal: 18})
	polygonMainnetClient := ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.PolygonMainnet, Decimal: 18})
	bnbClient := ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.Bnb, Decimal: 18})
	solanaClient := solana.NewSolanaClient(app.Cfg.Blockchain.SolanaRPC)
	bitcoinClient := bitcoin.NewBitcoinClient(app.Cfg.Blockchain.Bitcoin.Host, app.Cfg.Blockchain.Bitcoin.User, app.Cfg.Blockchain.Bitcoin.Pass)
	tronClient := tron.NewTronClient(app.Cfg.Blockchain.TronGRPC, app.Cfg.Blockchain.TronAPIKey)

	priceCache := core.NewPriceCache(redisClient)

	priceFetcher := core.NewPriceFetcher(core.PriceFetcherDeps{
		CoinGeckoClient: coingeckoClient,
		AlchemyClient:   alchemyClient,
		PriceCache:      priceCache,
		Repo:            &priceRepo,
		AllCoinInterval: app.Cfg.CoinGecko.PriceFetcher,
	})

	tokenRegistry := core.DefaultTokenRegistry(app.Cfg.TokenRegistry)
	priceService := core.NewPriceService(redisClient, &priceRepo, priceFetcher, priceCache, tokenRegistry)
	blockchainService := core.NewBlockchainService(core.BlockchainServiceDeps{
		EthMainnet:    ethMainnetClient,
		EthArbitrum:   ethArbitrumClient,
		EthBase:       ethBaseClient,
		Polygon:       polygonMainnetClient,
		BNB:           bnbClient,
		SOL:           solanaClient,
		BTC:           bitcoinClient,
		Tron:          tronClient,
		WalletRepo:    walletRepo,
		TokenRegistry: tokenRegistry,
		// Cache:         priceCache,
		Cache: core.NewNoOpCache(),
	})

	if err := blockchainService.ConnectAll(ctx); err != nil {
		logrus.WithError(err).Warn("failed to connect all blockchain clients")
	}

	walletService := core.NewWalletService(core.WalletDeps{
		WalletRepo: walletRepo, PriceService: priceService,
		UserRepo: userRepo, BlockchainService: blockchainService, TokenRegistry: tokenRegistry,
	})

	go priceFetcher.StartCoinFetcher(ctx)

	verifier, err := middleware.NewTokenVerifier(ctx, app.Cfg.Authorization.IssuerURL, app.Cfg.Authorization.ClientID)
	if err != nil {
		logrus.Fatalf("failed to create jwt verifier %s", err.Error())
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: false,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	httpAddr := fmt.Sprintf(":%d", app.Cfg.Server.PortHTTP)
	grpcAddr := fmt.Sprintf(":%d", app.Cfg.Server.PortGRPC)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryAuthInterceptor(verifier)))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	reflection.Register(grpcServer)

	priceGrpcServer := handlers.NewPriceGrpcHandler(priceService)
	walletGrpcServer := handlers.NewWalletGrpcHandler(walletService)

	pricev1.RegisterPriceServiceServer(grpcServer, priceGrpcServer)
	walletv1.RegisterWalletServiceServer(grpcServer, walletGrpcServer)

	if err := pricev1.RegisterPriceServiceHandlerFromEndpoint(ctx, gwmux, grpcAddr, opts); err != nil {
		logrus.Fatalf("failed to register price gateway: %v", err)
	}
	if err := walletv1.RegisterWalletServiceHandlerFromEndpoint(ctx, gwmux, grpcAddr, opts); err != nil {
		logrus.Fatalf("failed to register wallet gateway: %v", err)
	}

	httpHandler := setupHttpHandler(gwmux)

	runHttp(g, "http", httpAddr, httpHandler)
	runGrpc(g, "grpc", grpcAddr, grpcServer)

	<-ctx.Done()
	logrus.Info("Shutdown signal received, shutting down...")

	if err := g.Wait(); err != nil {
		logrus.Error("Stopped", "erro", err)
	} else {
		logrus.Info("Stopped cleanly")
	}

	logrus.Info("Server shutdown complete")
}

func runHttp(g *errgroup.Group, name string, addr string, handler http.Handler) {
	g.Go(func() error {
		srv := &http.Server{
			Addr:    addr,
			Handler: handler,
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server failed: %w %s", err, name)
		}
		return nil
	})
}

func runGrpc(g *errgroup.Group, name string, addr string, grpcServer *grpc.Server) {
	g.Go(func() error {
		listen, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on gRPC: %s %w %s", addr, err, name)
		}
		if err := grpcServer.Serve(listen); err != nil {
			return fmt.Errorf("gRPC server failed: %w %s", err, name)
		}
		return nil
	})
}

func setupHttpHandler(gwmux *runtime.ServeMux) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", gwmux)
	mux.Handle("/docs/", http.StripPrefix("/docs", docs.Handler()))

	return corsMiddleware(mux)
}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}
