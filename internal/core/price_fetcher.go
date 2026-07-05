package core

import (
	"context"
	"time"
	"tracker/internal/cache"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/coingecko"
	"tracker/internal/core/entity"
	"tracker/internal/db/models"
	repositories "tracker/internal/db/repo"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type PriceFetcher struct {
	coingeckoClient *coingecko.CoinGeckoClient
	alchemyClient   *alchemy.AlchemyClient
	cache           *cache.RedisClient
	repo            *repositories.PriceRepository
	interval        time.Duration
}

func NewPriceFetcher(
	coingeckoClient *coingecko.CoinGeckoClient,
	alchemyClient *alchemy.AlchemyClient,
	cache *cache.RedisClient,
	repo *repositories.PriceRepository,
	interval time.Duration,
) *PriceFetcher {
	return &PriceFetcher{
		coingeckoClient: coingeckoClient,
		alchemyClient:   alchemyClient,
		cache:           cache,
		repo:            repo,
		interval:        interval,
	}
}

func (f *PriceFetcher) Start(ctx context.Context) {
	logrus.Info("Starting price fetcher background worker...")

	f.pefrom(ctx)

	// Then run every interval
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Price fetcher stopped")
			return
		case <-ticker.C:
			f.pefrom(ctx)
		}
	}
}

// Solana: github.com/gagliardetto/solana-go (This is the most popular, active, and rock-solid Go library for Solana, maintained much better than the foundation's old one).

// Ethereum / EVM Chains (BNB, Arbitrum, Base, Polygon): github.com/ethereum/go-ethereum (The official go-ethereum / Geth client. This handles ETH and every single EVM-compatible chain using the exact same code layout).

// Bitcoin: github.com/btcsuite/btcd/rpcclient (The legendary, rock-solid Go implementation for Bitcoin RPC).

// TRON: github.com/fbsobreira/gotron-sdk (The standard Go client package for TRON).

// rubblelabs/ripple
// blinklabs-io/gouroboros

func (f *PriceFetcher) pefrom(ctx context.Context) {
	coins, err := f.coingeckoClient.GetCoins(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch coins")
		return
	}
	symbols := []string{}
	for _, coin := range coins {
		symbols = append(symbols, coin.Symbol)
	}
	prices := []entity.Price{}
	chunks := lo.Chunk(symbols, 25)
	for i := range chunks {
		chunk := chunks[i]
		prices, err = f.alchemyClient.GetPrices(ctx, chunk)
		if err != nil {
			logrus.WithError(err).WithField("chunk", chunk).Warn("Failed to fetch price chunk")
			continue
		}
	}
	f.storeCoinSnapshot(ctx, coins)
	f.storePriceSnapshot(ctx, prices)
}

func (f *PriceFetcher) storeCoinSnapshot(ctx context.Context, coins []coingecko.CoinGeckoCoin) {
	logrus.WithField("count", len(coins)).Debug("Storing snapshots...")

	var snapshots []models.CoinSnapshot
	for _, coin := range coins {
		snapshot := models.CoinSnapshot{
			ID:          uuid.New(),
			CoinID:      coin.ID,
			Symbol:      coin.Symbol,
			Name:        coin.Name,
			ImageURL:    coin.Image,
			LastUpdated: coin.LastUpdated,
			SnapshotAt:  time.Now(),
		}
		snapshots = append(snapshots, snapshot)
	}

	// Store in Redis cache
	if err := f.cache.SetJSON(ctx, "coins:all", snapshots, 60); err != nil {
		logrus.WithError(err).Error("Failed to cache in Redis")
	}
	if err := f.repo.SetCoinSnapshot(ctx, snapshots); err != nil {
		logrus.WithError(err).Error("Failed to store snapshots")
	}
}

func (f *PriceFetcher) storePriceSnapshot(ctx context.Context, prices []entity.Price) {
	var snapshots []models.PriceSnapshot
	for _, v := range prices {
		snapshots = append(snapshots, models.PriceSnapshot{
			ID:          uuid.New(),
			Symbol:      v.Symbol,
			PriceUSD:    v.PriceUSD,
			LastUpdated: v.LastUpdated,
		})
	}
	// Store in Redis cache
	if err := f.cache.SetJSON(ctx, "prices:all", snapshots, 60); err != nil {
		logrus.WithError(err).Error("Failed to cache in Redis")
	}
	if err := f.repo.SetPriceSnapshot(ctx, snapshots); err != nil {
		logrus.WithError(err).Error("Failed to store snapshots")
	}
}
