package core

import (
	"context"
	"fmt"
	"time"
	"tracker/internal/cache"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/coingecko"
	"tracker/internal/db/models"
	repositories "tracker/internal/db/repo"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PriceFetcher struct {
	coingeckoClient    *coingecko.CoinGeckoClient
	alchemyClient      *alchemy.AlchemyClient
	cache              *cache.RedisClient
	priceCache         *PriceCache
	repo               *repositories.PriceRepository
	allCoinInterval    time.Duration
	activeCoinInterval time.Duration
	log                *logrus.Entry
}

func NewPriceFetcher(
	coingeckoClient *coingecko.CoinGeckoClient,
	alchemyClient *alchemy.AlchemyClient,
	cache *cache.RedisClient,
	repo *repositories.PriceRepository,
	priceCache *PriceCache,
	allCoinInterval time.Duration,
	activeCoinInterval time.Duration,
) *PriceFetcher {
	return &PriceFetcher{
		coingeckoClient:    coingeckoClient,
		alchemyClient:      alchemyClient,
		cache:              cache,
		priceCache:         priceCache,
		repo:               repo,
		allCoinInterval:    allCoinInterval,
		activeCoinInterval: activeCoinInterval,
		log:                logrus.WithField("component", "PriceFetcher"),
	}
}

func (f *PriceFetcher) StartCoinFetcher(ctx context.Context) {
	fetch := func() {
		coins, err := f.coingeckoClient.GetCoins(ctx)
		if err != nil {
			f.log.WithError(err).Error("Failed to fetch coins")
			return
		}
		f.storeCoinSnapshot(ctx, coins)
	}
	fetch()

	ticker := time.NewTicker(f.allCoinInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.log.Info("fetcher stopped")
			return
		case <-ticker.C:
			f.log.Info("fetcher update")
			fetch()
		}
	}
}

func (f *PriceFetcher) StartActiveCoinFetcher(ctx context.Context) {
	fetch := func() {
		pricesToFetch := f.priceCache.GetPricesToWatch(ctx)
		if len(pricesToFetch) > 0 {
			prices := []models.Price{}
			coins, _ := f.priceCache.GetCoinsBySymbol(ctx, pricesToFetch)
			for _, coin := range coins {
				price, err := f.coingeckoClient.GetPrice(ctx, coin.CoinID)
				if err != nil {
					f.log.WithError(err).Error("Failed to fetch price")
					return
				}
				prices = append(prices, priceFromGecko(price))
			}
			f.updatePriceSnapshot(ctx, prices)
		}
	}
	fetch()

	ticker := time.NewTicker(f.activeCoinInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.log.Info("fetcher active prices stopped")
			return
		case <-ticker.C:
			f.log.Info("fetcher active prices update")
			fetch()
		}
	}
}

// Solana: github.com/gagliardetto/solana-go

// Ethereum / EVM Chains (BNB, Arbitrum, Base, Polygon): github.com/ethereum/go-ethereum

// Bitcoin: github.com/btcsuite/btcd/rpcclient

// TRON: github.com/fbsobreira/gotron-sdk

// rubblelabs/ripple
// blinklabs-io/gouroboros

func (f *PriceFetcher) storeCoinSnapshot(ctx context.Context, coins []coingecko.CoinGeckoCoin) {
	var snapshots []models.Coin
	for _, coin := range coins {
		snapshot := models.Coin{
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
	if err := f.priceCache.SetCoins(ctx, snapshots); err != nil {
		f.log.WithError(err).Error("Failed to cache in Redis")
	}
	// Store in Redis one by one
	for _, coin := range snapshots {
		if err := f.cache.SetJSON(ctx, fmt.Sprintf("coins:%s", coin.Symbol), coin, 1*time.Hour); err != nil {
			f.log.WithError(err).Error("Failed to cache in Redis")
		}
	}

	if err := f.repo.SetCoinSnapshot(ctx, snapshots); err != nil {
		f.log.WithError(err).Error("Failed to store snapshots")
	}
}

func (f *PriceFetcher) updatePriceSnapshot(ctx context.Context, prices []models.Price) {
	var snapshots []models.Price
	for _, price := range prices {
		snapshots = append(snapshots, models.Price{
			ID:          uuid.New(),
			Symbol:      price.Symbol,
			PriceUSD:    price.PriceUSD,
			LastUpdated: price.LastUpdated,
		})
		if err := f.priceCache.SetPrice(ctx, price.Symbol, price); err != nil {
			f.log.WithError(err).Error("Failed to cache in Redis")
		}
	}
	if err := f.repo.SetPriceSnapshot(ctx, snapshots); err != nil {
		f.log.WithError(err).Error("Failed to store snapshots")
	}
}

func priceFromGecko(price coingecko.CoinGeckoPrice) models.Price {
	return models.Price{
		CoinID:      price.ID,
		Name:        price.Name,
		Symbol:      price.Symbol,
		PriceUSD:    price.MarketData.CurrentPrice["usd"],
		LastUpdated: price.LastUpdated,
	}
}

// func (f *PriceFetcher) pefrom(ctx context.Context) {
// 	prices := []models.Price{}
// 	chunks := lo.Chunk(symbols, 25)
// 	for i := range chunks {
// 		chunk := chunks[i]
// 		prices, err = f.alchemyClient.GetPrices(ctx, chunk)
// 		if err != nil {
// 			f.log.WithError(err).WithField("chunk", chunk).Warn("Failed to fetch price chunk")
// 			continue
// 		}
// 	}
// 	f.storePriceSnapshot(ctx, prices)
// }
