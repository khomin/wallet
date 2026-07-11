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
	handlers "tracker/internal/api/handles"
	"tracker/internal/cache"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/coingecko"
	"tracker/internal/core"
	"tracker/internal/db"
	repositories "tracker/internal/db/repo"

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

	coingeckoClient := coingecko.NewCoinGeckoClient(app.Cfg.CoinGecko.APIKey)
	alchemyClient := alchemy.NewAlchemyClient(app.Cfg.Alchemy.APIKey)
	priceRepo := repositories.NewPriceRepository(db)

	// Create a context that can be cancelled for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	priceCache := core.NewPriceCache(redisClient)

	// Create the price fetcher with 60 second interval
	priceFetcher := core.NewPriceFetcher(
		coingeckoClient,
		alchemyClient,
		priceRepo,
		priceCache,
		60*time.Second,
		10*time.Second,
	)

	priceService := core.NewPriceService(redisClient, priceRepo, priceFetcher, priceCache)
	priceHandler := handlers.NewPriceHandler(priceService)

	go priceFetcher.StartCoinFetcher(ctx)

	r := gin.Default()

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/coins", priceHandler.GetCoins)
		v1.GET("/coins/:id", priceHandler.GetCoin)
		v1.GET("/prices", priceHandler.GetPrices)
		v1.GET("/prices/:id", priceHandler.GetPrice)
		v1.GET("/wallets", walletHandler.ListWallets)
		v1.POST("/wallets", walletHandler.AddWallet)
	}
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

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

// TODO: handle API from clients
//  - add wallet, delete wallet, get wallet/s
//  - device authentification, tokens
//  - profile
