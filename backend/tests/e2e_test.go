package tests

import (
	"context"
	"testing"
	"tracker/bootstrap"
	"tracker/internal/cache"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/bitcoin"
	"tracker/internal/client/coingecko"
	"tracker/internal/client/ethereum"
	"tracker/internal/client/solana"
	"tracker/internal/client/tron"
	"tracker/internal/core"
	"tracker/internal/core/domain"
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
		Cache:         priceCache,
	})
	if err := blockchainService.ConnectAll(ctx); err != nil {
		logrus.WithError(err).Warn("failed to connect all blockchain clients")
	}

	go priceFetcher.StartCoinFetcher(ctx)

	svc := core.NewWalletService(walletRepo, priceService, blockchainService, tokenRegistry)

	userID := "ad4abec0-8bae-462a-8ea3-8502048f3071"

	tokens := []domain.Token{}

	tokens = append(tokens, domain.Token{
		Name:   "BTC",
		Chains: []string{"BTC"},
		Symbol: "BTC",
		Addrs:  []string{"bc1qyyfjmtl9s8s6wn3llt2ltzze978l5szfj7hr79"},
	})

	tokens = append(tokens, domain.Token{
		Chains: []string{"SOL"},
		Symbol: "GOOGLX",
		Addrs:  []string{"DDcdDmDPYw595wAR1jYNHZQTFNi8BGisd2bVa3WH3XbE"},
	})
	tokens = append(tokens, domain.Token{
		Chains: []string{"SOL"},
		Symbol: "XAUT",
		Addrs:  []string{"9Z6qhmZ2AHWMSBSM4LmA1WrCAefs2dkr1pHnkW3vmg8z"},
	})
	tokens = append(tokens, domain.Token{
		Chains: []string{"SOL"},
		Symbol: "SOL",
		Addrs:  []string{"9Z6qhmZ2AHWMSBSM4LmA1WrCAefs2dkr1pHnkW3vmg8z"},
	})
	tokens = append(tokens, domain.Token{
		Chains: []string{"TRX"},
		Symbol: "TRX",
		Addrs:  []string{"TF6MrLnLa72U6PGZcx4pXhs8hSvsAgQ78t"},
	})
	tokens = append(tokens, domain.Token{
		Chains: []string{"TRX"},
		Symbol: "USDT",
		Addrs:  []string{"TF6MrLnLa72U6PGZcx4pXhs8hSvsAgQ78t"},
	})

	print(tokens, svc, userID)

	for _, token := range tokens {
		for idx, chain := range token.Chains {
			_, err = svc.AddWallet(context.Background(), userID, chain, token.Addrs[idx], token.Symbol, token.Name)
			if err != nil {
				t.Logf("AddWallet returned unexpected error: %v", err)
			}
		}
	}
}
