package core

import (
	"context"
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
		coinsMarket, err := f.coingeckoClient.GetMarket(ctx)
		if err != nil {
			f.log.WithError(err).Error("Failed to fetch coins")
			return
		}
		coins := f.fromGeckoToCoin(coinsMarket)
		prices := f.fromGeckoToCoinPrice(coinsMarket)

		if err := f.priceCache.SetCoins(ctx, coins); err != nil {
			f.log.WithError(err).Error("Failed to cache coins")
		}
		if err := f.priceCache.SetPrices(ctx, prices); err != nil {
			f.log.WithError(err).Error("Failed to cache prices")
		}
		if err := f.repo.SetCoinSnapshot(ctx, coins); err != nil {
			f.log.WithError(err).Error("Failed to store snapshots")
		}
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

// Solana: github.com/gagliardetto/solana-go

// Ethereum / EVM Chains (BNB, Arbitrum, Base, Polygon): github.com/ethereum/go-ethereum

// Bitcoin: github.com/btcsuite/btcd/rpcclient

// TRON: github.com/fbsobreira/gotron-sdk

// rubblelabs/ripple
// blinklabs-io/gouroboros

func (f *PriceFetcher) updatePriceSnapshot(ctx context.Context, prices []models.CoinPrice) {
	var snapshots []models.CoinPrice
	for _, price := range prices {
		snapshots = append(snapshots, models.CoinPrice{
			ID:           uuid.New(),
			Symbol:       price.Symbol,
			CurrentPrice: price.CurrentPrice,
			LastUpdated:  price.LastUpdated,
		})
		if err := f.priceCache.SetPrice(ctx, price.Symbol, price); err != nil {
			f.log.WithError(err).Error("Failed to cache in Redis")
		}
	}
	if err := f.repo.SetPriceSnapshot(ctx, snapshots); err != nil {
		f.log.WithError(err).Error("Failed to store snapshots")
	}
}

func (f *PriceFetcher) fromGeckoToCoin(price []coingecko.CoinGeckoCoin) []models.Coin {
	res := []models.Coin{}
	for _, i := range price {
		res = append(res, models.Coin{
			CoinID:      i.ID,
			Name:        i.Name,
			Symbol:      i.Symbol,
			ImageURL:    i.Image,
			LastUpdated: i.LastUpdated,
		})
	}
	return res
}

func (f *PriceFetcher) fromGeckoToCoinPrice(price []coingecko.CoinGeckoCoin) []models.CoinPrice {
	res := []models.CoinPrice{}
	for _, i := range price {
		res = append(res, models.CoinPrice{
			CoinID:                         i.ID,
			Name:                           i.Name,
			Symbol:                         i.Symbol,
			CurrentPrice:                   i.CurrentPrice,
			Change_24h:                     i.PriceChange24h,
			MarketCap:                      i.MarketCap,
			TotalVolume:                    i.TotalVolume,
			High_24h:                       i.High24h,
			Low_24h:                        i.Low24h,
			PriceChange_24h:                i.PriceChange24h,
			PriceChangePercentage_24h:      i.PriceChangePercent24h,
			MarketCapChange_24h:            i.MarketCapChange24h,
			MarketCapChange_percentage_24h: i.MarketCapChangePercent24h,
			LastUpdated:                    i.LastUpdated,
		})
	}
	return res
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
