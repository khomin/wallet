package core

import (
	"context"
	"fmt"
	"time"
	"tracker/internal/cache"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/coingecko"
	"tracker/internal/core/entity"
	"tracker/internal/db/models"
	repositories "tracker/internal/db/repo"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PriceFetcher struct {
	coingeckoClient    *coingecko.CoinGeckoClient
	alchemyClient      *alchemy.AlchemyClient
	cache              *cache.RedisClient
	repo               *repositories.PriceRepository
	allCoinInterval    time.Duration
	activeCoinInterval time.Duration
}

func NewPriceFetcher(
	coingeckoClient *coingecko.CoinGeckoClient,
	alchemyClient *alchemy.AlchemyClient,
	cache *cache.RedisClient,
	repo *repositories.PriceRepository,
	allCoinInterval time.Duration,
	activeCoinInterval time.Duration,
) *PriceFetcher {
	return &PriceFetcher{
		coingeckoClient:    coingeckoClient,
		alchemyClient:      alchemyClient,
		cache:              cache,
		repo:               repo,
		allCoinInterval:    allCoinInterval,
		activeCoinInterval: activeCoinInterval,
	}
}

func (f *PriceFetcher) StartCoinFetcher(ctx context.Context) {
	log := logrus.WithField("PriceFetcher", "StartCoinFetcher")
	fetch := func() {
		coins, err := f.coingeckoClient.GetCoins(ctx)
		if err != nil {
			logrus.WithError(err).Error("Failed to fetch coins")
			return
		}
		f.storeCoinSnapshot(ctx, coins)
	}
	log.Info("fetcher update initial")
	fetch()

	ticker := time.NewTicker(f.allCoinInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("fetcher stopped")
			return
		case <-ticker.C:
			log.Info("fetcher update")
			fetch()
		}
	}
}

func (f *PriceFetcher) StartActiveCoinFetcher(ctx context.Context) {
	log := logrus.WithField("PriceFetcher", "StartActiveCoinFetcher")
	fetch := func() {
		pricesToUpdate := f.getActivePrices(ctx)
		prices := []entity.Price{}
		if len(pricesToUpdate) > 0 {
			for _, symbol := range pricesToUpdate {
				price, err := f.coingeckoClient.GetPrice(ctx, symbol)
				if err != nil {
					logrus.WithError(err).Error("Failed to fetch price")
					return
				}
				prices = append(prices, entity.Price{
					ID:       price.ID,
					Symbol:   price.Symbol,
					PriceUSD: price.MarketData.CurrentPrice["usd"],
					// Change24h: price.MarketData.High24h[],
					LastUpdated: price.LastUpdated,
				})
			}
			f.updatePriceSnapshot(ctx, prices)
		}
	}
	log.Info("fetcher update initial")
	fetch()

	ticker := time.NewTicker(f.activeCoinInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("fetcher stopped")
			return
		case <-ticker.C:
			log.Info("fetcher update")
			fetch()
		}
	}
}

func (f *PriceFetcher) getActivePrices(ctx context.Context) []string {
	log := logrus.WithField("PriceFetcher", "getActivePrices")
	var symbols []string
	if err := f.cache.Get(ctx, "active-price", &symbols); err != nil {
		log.Info("no active coins to fetch info")
		return symbols
	}
	return symbols
}

// Solana: github.com/gagliardetto/solana-go

// Ethereum / EVM Chains (BNB, Arbitrum, Base, Polygon): github.com/ethereum/go-ethereum

// Bitcoin: github.com/btcsuite/btcd/rpcclient

// TRON: github.com/fbsobreira/gotron-sdk

// rubblelabs/ripple
// blinklabs-io/gouroboros

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

	// Store in Redis cache all
	if err := f.cache.SetJSON(ctx, "coins:all", snapshots, 1*time.Hour); err != nil {
		logrus.WithError(err).Error("Failed to cache in Redis")
	}
	// Store in Redis one by one
	for _, coin := range snapshots {
		if err := f.cache.SetJSON(ctx, fmt.Sprintf("coins:%s", coin.Symbol), coin, 1*time.Hour); err != nil {
			logrus.WithError(err).Error("Failed to cache in Redis")
		}
	}

	if err := f.repo.SetCoinSnapshot(ctx, snapshots); err != nil {
		logrus.WithError(err).Error("Failed to store snapshots")
	}
}

func (f *PriceFetcher) updatePriceSnapshot(ctx context.Context, prices []entity.Price) {
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
	if err := f.cache.SetJSON(ctx, "prices:all", snapshots, 60*time.Second); err != nil {
		logrus.WithError(err).Error("Failed to cache in Redis")
	}
	if err := f.repo.SetPriceSnapshot(ctx, snapshots); err != nil {
		logrus.WithError(err).Error("Failed to store snapshots")
	}
}

// func (f *PriceFetcher) pefrom(ctx context.Context) {
// 	prices := []entity.Price{}
// 	chunks := lo.Chunk(symbols, 25)
// 	for i := range chunks {
// 		chunk := chunks[i]
// 		prices, err = f.alchemyClient.GetPrices(ctx, chunk)
// 		if err != nil {
// 			logrus.WithError(err).WithField("chunk", chunk).Warn("Failed to fetch price chunk")
// 			continue
// 		}
// 	}
// 	f.storePriceSnapshot(ctx, prices)
// }
