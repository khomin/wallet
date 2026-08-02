package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
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

	// Create a context that can be cancelled for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	priceCache := core.NewPriceCache(redisClient)

	// Create the price fetcher with 60 second interval
	priceFetcher := core.NewPriceFetcher(core.PriceFetcherDeps{
		CoinGeckoClient: coingeckoClient,
		AlchemyClient:   alchemyClient,
		PriceCache:      priceCache,
		Repo:            &priceRepo,
		AllCoinInterval: app.Cfg.CoinGecko.PriceFetcher,
	})

	tokenRegistry := core.DefaultTokenRegistry(app.Cfg.TokenRegistry)
	priceService := core.NewPriceService(redisClient, &priceRepo, priceFetcher, priceCache, tokenRegistry)
	walletRepo := repositories.NewWalletRepository(db)
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
	// assetsHandler := handlers.NewPriceHandler(priceService)

	if err := blockchainService.ConnectAll(ctx); err != nil {
		logrus.WithError(err).Warn("failed to connect all blockchain clients")
	}

	walletService := core.NewWalletService(walletRepo, priceService, blockchainService, tokenRegistry)
	walletHandler := handlers.NewWalletHandler(walletService)

	go priceFetcher.StartCoinFetcher(ctx)

	verifier, err := middleware.NewTokenVerifier(ctx, app.Cfg.Authorization.IssuerURL, app.Cfg.Authorization.ClientID)
	if err != nil {
		logrus.Fatalf("failed to create jwt verifier %s", err.Error())
	}

	r := gin.Default()

	// CORS – allow Vite dev server (localhost:5173) and any other origins you need
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// v1 := r.Group("/api/v1")
	// {
	// 	protected := v1.Group("").Use(middleware.Auth(verifier))
	// 	v1.GET("/coins", assetsHandler.GetCoins)
	// 	v1.GET("/coins/:id", assetsHandler.GetCoin)
	// 	v1.GET("/coins/search", assetsHandler.SearchCoin)
	// 	v1.GET("/prices", assetsHandler.GetPrices)
	// 	v1.GET("/prices/:id", assetsHandler.GetPrice)
	// 	protected.GET("/wallets/:id", walletHandler.GetWallet)
	// 	protected.GET("/wallets", walletHandler.ListWallets)
	// 	protected.POST("/wallets", walletHandler.AddWallet)
	// 	protected.PUT("/wallets", walletHandler.EditWallet)
	// 	protected.DELETE("/wallets", walletHandler.DeleteWallet)
	// }
	// r.GET("/health", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status":    "ok",
	// 		"timestamp": time.Now().UTC().Format(time.RFC3339),
	// 	})
	// })
	// print(assetsHandler, verifier, walletHandler)
	print(verifier, walletHandler)

	httpAddr := fmt.Sprintf(":%d", app.Cfg.Server.PortHTTP)
	grpcPort := fmt.Sprintf(":%d", app.Cfg.Server.PortGRPC)
	gwmux := runtime.NewServeMux()

	grpcServer := grpc.NewServer(grpc.EmptyServerOption{})
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logrus.Fatalf("failed to listen on gRPC port %s: %v", grpcPort, err)
	}
	httpHandler := setupHttpHandler(gwmux)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: httpHandler,
	}

	walletGrpcServer := handlers.NewWalletGrpcHandler(walletService)
	priceGrpcServer := handlers.NewPriceGrpcHandler(priceService)

	walletv1.RegisterWalletServiceServer(grpcServer, walletGrpcServer)
	pricev1.RegisterPriceServiceServer(grpcServer, priceGrpcServer)

	walletv1.RegisterWalletServiceHandlerServer(ctx, gwmux, walletGrpcServer)
	pricev1.RegisterPriceServiceHandlerServer(ctx, gwmux, priceGrpcServer)

	reflection.Register(grpcServer)

	go func() {
		logrus.WithField("port", grpcPort).Info("Starting gRPC server")
		if err := grpcServer.Serve(lis); err != nil {
			logrus.WithError(err).Fatal("gRPC server failed")
		}
	}()
	go func() {
		logrus.WithField("port", app.Cfg.Server.PortHTTP).Info("HTTP server started")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal (Ctrl+C or kill)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down gracefully...")
	cancel()

	// Shutdown HTTP server with 5 second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Error("HTTP server forced to shutdown")
	}

	logrus.Info("Server shutdown complete")
}

func setupHttpHandler(gwmux *runtime.ServeMux) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", gwmux)

	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gen/openapiv2/v1/wallet.swagger.json")
	})

	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Wallet API Docs</title>
			<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
		</head>
		<body>
			<div id="swagger-ui"></div>
			<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
			<script>
				SwaggerUIBundle({
					url: '/swagger.json',
					dom_id: '#swagger-ui',
				});
			</script>
		</body>
		</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

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
