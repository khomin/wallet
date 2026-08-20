package cmd

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"tracker/bootstrap"
	"tracker/internal/cache"
	"tracker/internal/client"
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

func TestGenerateWallets(t *testing.T) {
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
	alertRepo := repositories.NewAlertRepository(db)

	coingeckoClient := coingecko.NewCoinGeckoClient(app.Cfg.CoinGecko.APIKey)
	alchemyClient := alchemy.NewAlchemyClient(app.Cfg.Alchemy.APIKey)

	ethMainnetClient := client.NewRateLimiterProvider(&app.Cfg.Blockchain.RateLimitConfig, ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.EthereumMainnet, Decimal: 18}))
	ethArbitrumClient := client.NewRateLimiterProvider(&app.Cfg.Blockchain.RateLimitConfig, ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.EthereumArbitrum, Decimal: 18}))
	ethBaseClient := client.NewRateLimiterProvider(&app.Cfg.Blockchain.RateLimitConfig, ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.EthereumBase, Decimal: 18}))
	polygonMainnetClient := client.NewRateLimiterProvider(&app.Cfg.Blockchain.RateLimitConfig, ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.PolygonMainnet, Decimal: 18}))
	bnbClient := client.NewRateLimiterProvider(&app.Cfg.Blockchain.RateLimitConfig, ethereum.NewEthereumClient(ethereum.EthConfig{RpcURL: app.Cfg.Blockchain.Bnb, Decimal: 18}))
	solanaClient := client.NewRateLimiterProvider(&app.Cfg.Blockchain.RateLimitConfig, solana.NewSolanaClient(app.Cfg.Blockchain.SolanaRPC))
	bitcoinClient := client.NewRateLimiterProvider(&app.Cfg.Blockchain.RateLimitConfig, bitcoin.NewBitcoinClient(app.Cfg.Blockchain.Bitcoin.Host, app.Cfg.Blockchain.Bitcoin.User, app.Cfg.Blockchain.Bitcoin.Pass))
	tronClient := client.NewRateLimiterProvider(&app.Cfg.Blockchain.RateLimitConfig, tron.NewTronClient(app.Cfg.Blockchain.TronGRPC, app.Cfg.Blockchain.TronAPIKey))

	// Create a context that can be cancelled for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	priceCache := core.NewPriceCache(redisClient)

	// Create the price fetcher with 60 second interval
	priceFetcher := core.NewPriceFetcher(core.PriceFetcherDeps{
		CoinGeckoClient:    coingeckoClient,
		AlchemyClient:      alchemyClient,
		PriceCache:         priceCache,
		PriceRepo:          priceRepo,
		AlertRepo:          alertRepo,
		FetchCoinsInterval: app.Cfg.CoinGecko.PriceFetcher,
	})

	tokenRegistry := core.DefaultTokenRegistry(app.Cfg.TokenRegistry)
	priceService := core.NewPriceService(redisClient, priceRepo, priceFetcher, priceCache, tokenRegistry)
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

	go priceFetcher.StartFetcher(ctx)

	svc := core.NewWalletService(core.WalletDeps{
		WalletRepo:        walletRepo,
		PriceService:      priceService,
		UserRepo:          repositories.NewUserRepo(db),
		BlockchainService: blockchainService,
		TokenRegistry:     tokenRegistry,
	})

	// import known wallets
	file, err := os.ReadFile("../../resources/known_wallets.json")
	if err != nil {
		logrus.Fatal(err)
	}
	var imported KnownWallets
	json.Unmarshal(file, &imported)

	// import rich wallets
	// eth
	file, err = os.ReadFile("./wallets/ETHEREUM/rich_01.txt")
	if err != nil {
		logrus.Fatal(err)
	}
	fileStr := string(file)
	lines := strings.Split(fileStr, "\n")
	for _, v := range lines {
		imported.Tokens = append(imported.Tokens, KnownToken{
			Chain:  "ETH",
			Symbol: "ETH",
			Addr:   v,
		})
	}

	// trx
	file, err = os.ReadFile("./wallets/TRON/TetherUSDT_TRC20_Rich_Address_Balance.txt")
	if err != nil {
		logrus.Fatal(err)
	}
	fileStr = string(file)
	lines = strings.Split(fileStr, "\n")
	for _, v := range lines {
		v := strings.Split(v, ",")[0]
		imported.Tokens = append(imported.Tokens, KnownToken{
			Chain:  "TRX",
			Symbol: "TRX",
			Addr:   v,
		})
	}

	// add in loop
	for _, token := range imported.Tokens {
		if err = svc.CreateWallet(context.Background(), domain.User{
			ID:    imported.UserID,
			Name:  imported.UserID,
			Email: imported.UserID,
		}, token.Chain, token.Addr, token.Symbol, token.Label); err != nil {
			logrus.Warnf("AddWallet returned error: %v, %s", err, token.Addr)
		} else {
			logrus.Infof("added wallet: %s: %s", token.Chain, token.Addr)
		}
	}
	logrus.Info("completed")
}
