package core

import (
	"context"
	"time"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/coingecko"
	"tracker/internal/core/domain"

	"github.com/sirupsen/logrus"
)

type EmailSender interface {
	Send(ctx context.Context, recipient, subject, body string) error
}

type PriceFetcherDeps struct {
	CoinGeckoClient    *coingecko.CoinGeckoClient
	AlchemyClient      *alchemy.AlchemyClient
	PriceCache         PriceCache
	PriceRepo          PriceRepository
	FetchCoinsInterval time.Duration
	AlertRepo          AlertRepository
	UserRepo           UserRepo
	AlertService       *AlertService
	EmailSender        EmailSender
}

type PriceFetcher struct {
	coingeckoClient    *coingecko.CoinGeckoClient
	alchemyClient      *alchemy.AlchemyClient
	priceCache         PriceCache
	priceRepo          PriceRepository
	alertRepo          AlertRepository
	userRepo           UserRepo
	emailSender        EmailSender
	fetchCoinsInterval time.Duration
	alertService       *AlertService
	log                *logrus.Entry
}

func NewPriceFetcher(deps PriceFetcherDeps) *PriceFetcher {
	return &PriceFetcher{
		coingeckoClient:    deps.CoinGeckoClient,
		alchemyClient:      deps.AlchemyClient,
		priceCache:         deps.PriceCache,
		priceRepo:          deps.PriceRepo,
		alertRepo:          deps.AlertRepo,
		userRepo:           deps.UserRepo,
		emailSender:        deps.EmailSender,
		fetchCoinsInterval: deps.FetchCoinsInterval,
		alertService:       deps.AlertService,
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
	_ = f.priceRepo.SetCoins(ctx, coins)
	_ = f.priceRepo.SetPrices(ctx, prices)

	go f.alertService.EvaluateAlerts(ctx)
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
