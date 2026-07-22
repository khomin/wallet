package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tracker/bootstrap"
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
	"github.com/sirupsen/logrus"
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

	ethMainnetClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.EthereumMainnet)
	ethArbitrumClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.EthereumArbitrum)
	ethBaseClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.EthereumBase)
	polygonMainnetClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.PolygonMainnet)
	bnbClient := ethereum.NewEthereumClient(app.Cfg.Blockchain.Bnb)
	solanaClient := solana.NewSolanaClient(app.Cfg.Blockchain.SolanaRPC)
	bitcoinClient := bitcoin.NewBitcoinClient(app.Cfg.Blockchain.BitcoinRPCHost, app.Cfg.Blockchain.BitcoinRPCUser, app.Cfg.Blockchain.BitcoinRPCPass)
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

	priceService := core.NewPriceService(redisClient, &priceRepo, priceFetcher, priceCache)
	priceHandler := handlers.NewPriceHandler(priceService)
	tokenRegistry := core.DefaultTokenRegistry(app.Cfg.TokenRegistry)
	walletRepo := repositories.NewWalletRepository(db)

	blockchainService := core.NewBlockchainService(
		ethMainnetClient, ethArbitrumClient, ethBaseClient, polygonMainnetClient, bnbClient,
		solanaClient, bitcoinClient, tronClient,
		walletRepo, tokenRegistry,
	)
	if err := blockchainService.ConnectAll(ctx); err != nil {
		logrus.WithError(err).Warn("failed to connect all blockchain clients")
	}

	walletService := core.NewWalletService(walletRepo, priceService, blockchainService)
	walletHandler := handlers.NewWalletHandler(walletService)

	go priceFetcher.StartCoinFetcher(ctx)

	verifier, err := middleware.NewTokenVerifier(ctx, app.Cfg.Authorization.IssuerURL, app.Cfg.Authorization.ClientID)
	if err != nil {
		logrus.Panicf("failed to create jwt verifier")
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

	v1 := r.Group("/api/v1")
	{
		v1.GET("/coins", priceHandler.GetCoins)
		v1.GET("/coins/:id", priceHandler.GetCoin)
		v1.GET("/prices", priceHandler.GetPrices)
		v1.GET("/prices/:id", priceHandler.GetPrice)
		// protected wallets
		protected := v1.Group("").Use(middleware.Auth(verifier))
		protected.GET("/wallets", walletHandler.ListWallets)
		protected.POST("/wallets", walletHandler.AddWallet)
		protected.PUT("/wallets", walletHandler.EditWallet)
		protected.GET("/wallets/balance", walletHandler.GetWalletBalance)
		protected.DELETE("/wallets", walletHandler.DeleteWallet)
	}
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// TODO: sdk
	// Bitcoin: github.com/btcsuite/btcd/rpcclient
	// TRON: github.com/fbsobreira/gotron-sdk
	// rubblelabs/ripple
	// blinklabs-io/gouroboros

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Cfg.Server.Port),
		Handler: r,
	}

	// Start server in a goroutine (non-blocking)
	go func() {
		logrus.WithField("port", app.Cfg.Server.Port).Info("HTTP server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// ============================================
	// 11. GRACEFUL SHUTDOWN
	// ============================================
	// Wait for interrupt signal (Ctrl+C or kill)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down gracefully...")

	// Signal the background worker to stop
	cancel()

	// Shutdown HTTP server with 5 second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Error("HTTP server forced to shutdown")
	}

	logrus.Info("Server shutdown complete")
}
