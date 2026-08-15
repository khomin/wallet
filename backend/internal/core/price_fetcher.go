package core

import (
	"context"
	"time"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/coingecko"
	"tracker/internal/core/domain"

	"github.com/sirupsen/logrus"
)

type PriceFetcherDeps struct {
	CoinGeckoClient    *coingecko.CoinGeckoClient
	AlchemyClient      *alchemy.AlchemyClient
	PriceCache         PriceCache
	PriceRepo          PriceRepository
	FetchCoinsInterval time.Duration
}

type PriceFetcher struct {
	coingeckoClient    *coingecko.CoinGeckoClient
	alchemyClient      *alchemy.AlchemyClient
	priceCache         PriceCache
	priceRepo          PriceRepository
	fetchCoinsInterval time.Duration
	log                *logrus.Entry
}

func NewPriceFetcher(deps PriceFetcherDeps) *PriceFetcher {
	return &PriceFetcher{
		coingeckoClient:    deps.CoinGeckoClient,
		alchemyClient:      deps.AlchemyClient,
		priceCache:         deps.PriceCache,
		priceRepo:          deps.PriceRepo,
		fetchCoinsInterval: deps.FetchCoinsInterval,
		log:                logrus.WithField("component", "PriceFetcher"),
	}
}

func (f *PriceFetcher) StartFetcher(ctx context.Context) {
	ticker := time.NewTicker(f.fetchCoinsInterval)
	defer ticker.Stop()

	f.fetch(ctx)
	for {
		select {
		case <-ctx.Done():
			f.log.Info("fetcher stopped")
			return
		case <-ticker.C:
			f.log.Info("fetcher update")
			f.fetch(ctx)
		}
	}
}

func (f *PriceFetcher) LoadCache(ctx context.Context) {
	coins, err := f.priceRepo.GetCoins(ctx)
	if err != nil {
		f.log.WithError(err).Error("Failed to read coin snapshot")
		return
	}
	prices, err := f.priceRepo.GetPrices(ctx)
	if err != nil {
		f.log.WithError(err).Error("Failed to read price snapshot")
		return
	}
	if err := f.priceCache.SetCoins(ctx, coins); err != nil {
		f.log.WithError(err).Error("Failed to cache coins")
	}
	if err := f.priceCache.SetPrices(ctx, prices); err != nil {
		f.log.WithError(err).Error("Failed to cache prices")
	}
}

func (f *PriceFetcher) fetch(ctx context.Context) {
	coinsMarket, err := f.coingeckoClient.GetMarket(ctx)
	if err != nil {
		f.log.WithError(err).Error("Failed to fetch coins from Gecko")
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
	if err := f.priceRepo.SetCoins(ctx, coins); err != nil {
		f.log.WithError(err).Error("Failed to store coin snapshots")
	}
	if err := f.priceRepo.SetPrices(ctx, prices); err != nil {
		f.log.WithError(err).Error("Failed to store price snapshots")
	}
	//
}

func (f *PriceFetcher) fromGeckoToCoin(prices []coingecko.CoinGeckoCoin) []domain.TokenID {
	res := make([]domain.TokenID, 0, len(prices))
	for _, i := range prices {
		res = append(res, domain.TokenID{
			ID:       i.ID,
			Name:     i.Name,
			Symbol:   i.Symbol,
			ImageURL: i.Image,
		})
	}
	return res
}

func (f *PriceFetcher) fromGeckoToCoinPrice(prices []coingecko.CoinGeckoCoin) []domain.TokenPrice {
	res := make([]domain.TokenPrice, len(prices))
	for i, p := range prices {
		res[i] = domain.TokenPrice{
			ID:                             p.ID,
			Name:                           p.Name,
			Symbol:                         p.Symbol,
			CurrentPrice:                   p.CurrentPrice,
			Change_24h:                     p.PriceChange24h,
			MarketCap:                      p.MarketCap,
			TotalVolume:                    p.TotalVolume,
			High_24h:                       p.High24h,
			Low_24h:                        p.Low24h,
			PriceChange_24h:                p.PriceChange24h,
			PriceChangePercentage_24h:      p.PriceChangePercent24h,
			MarketCapChange_24h:            p.MarketCapChange24h,
			MarketCapChange_percentage_24h: p.MarketCapChangePercent24h,
			UpdatedAt:                      p.LastUpdated,
		}
	}
	return res
}
