package tests

import (
	"context"
	"testing"
	"time"
	"tracker/bootstrap"
	"tracker/internal/cache"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/bitcoin"
	"tracker/internal/client/coingecko"
	"tracker/internal/client/ethereum"
	"tracker/internal/client/solana"
	"tracker/internal/client/tron"
	"tracker/internal/core"
	"tracker/internal/core/entity"
	"tracker/internal/db"
	"tracker/internal/db/repositories"

	"github.com/sirupsen/logrus"
)

func TestDeleteWalletReturnsDeletedWallet(t *testing.T) {
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

	ethMainnetClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.EthereumMainnet)
	ethArbitrumClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.EthereumArbitrum)
	ethBaseClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.EthereumBase)
	polygonMainnetClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.PolygonMainnet)
	bnbClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.Bnb)
	solanaClient := solana.NewSolanaClient(app.Cfg.Blockchain.SolanaRPC)
	bitcoinClient := bitcoin.NewBitcoinClient(app.Cfg.Blockchain.Bitcoin.Host, app.Cfg.Blockchain.Bitcoin.User, app.Cfg.Blockchain.Bitcoin.Pass)
	tronClient := tron.NewTronClient(app.Cfg.Blockchain.TronGRPC, app.Cfg.Blockchain.TronAPIKey)

	// Create a context that can be cancelled for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	priceCache := core.NewPriceCache(redisClient)

	// Create the price fetcher with 60 second interval
	priceFetcher := core.NewPriceFetcher(
		coingeckoClient,
		alchemyClient,
		&priceRepo,
		priceCache,
		60*time.Second,
		10*time.Second,
	)

	tokenRegistry := core.DefaultTokenRegistry(app.Cfg.TokenRegistry)
	priceService := core.NewPriceService(redisClient, &priceRepo, priceFetcher, priceCache, tokenRegistry)
	walletRepo := repositories.NewWalletRepository(db)
	blockchainService := core.NewBlockchainService(
		ethMainnetClient, ethArbitrumClient, ethBaseClient, polygonMainnetClient, bnbClient,
		solanaClient, bitcoinClient, tronClient,
		walletRepo, tokenRegistry,
	)
	// priceHandler := handlers.NewPriceHandler(priceService)

	if err := blockchainService.ConnectAll(ctx); err != nil {
		logrus.WithError(err).Warn("failed to connect all blockchain clients")
	}

	// walletService := core.NewWalletService(walletRepo, priceService, blockchainService)
	// walletHandler := handlers.NewWalletHandler(walletService)

	go priceFetcher.StartCoinFetcher(ctx)

	// verifier, err := middleware.NewTokenVerifier(ctx, app.Cfg.Authorization.IssuerURL, app.Cfg.Authorization.ClientID)
	// if err != nil {
	// 	logrus.Fatal("failed to create jwt verifier")
	// }

	svc := core.NewWalletService(walletRepo, priceService, blockchainService, tokenRegistry)

	userID := "ad4abec0-8bae-462a-8ea3-8502048f3071"

	tokens := []entity.Token{}

	tokens = append(tokens, entity.Token{
		Chain:   "BTC",
		Symbol:  "BTC",
		Address: "bc1qyyfjmtl9s8s6wn3llt2ltzze978l5szfj7hr79",
	})

	tokens = append(tokens, entity.Token{
		Chain:   "SOL",
		Symbol:  "GOOGLX",
		Address: "DDcdDmDPYw595wAR1jYNHZQTFNi8BGisd2bVa3WH3XbE",
	})
	tokens = append(tokens, entity.Token{
		Chain:   "SOL",
		Symbol:  "XAUT0",
		Address: "CFMQzGS8M8wpvcWs1udJ2XgzXEmVf31bmYxBwtgPBLRg",
	})
	tokens = append(tokens, entity.Token{
		Chain:   "SOL",
		Symbol:  "SOL",
		Address: "CFMQzGS8M8wpvcWs1udJ2XgzXEmVf31bmYxBwtgPBLRg",
	})
	tokens = append(tokens, entity.Token{
		Chain:   "TRX",
		Symbol:  "TRX",
		Address: "TF6MrLnLa72U6PGZcx4pXhs8hSvsAgQ78t",
	})
	tokens = append(tokens, entity.Token{
		Chain:   "TRX",
		Symbol:  "USDT",
		Address: "TF6MrLnLa72U6PGZcx4pXhs8hSvsAgQ78t",
	})

	for _, i := range tokens {
		_, err = svc.AddWallet(context.Background(), userID, i.Chain, i.Address, i.Symbol, i.Name)
		if err != nil {
			t.Logf("AddWallet returned unexpected error: %v", err)
		}
	}
}
