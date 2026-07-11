package core

import (
	"context"
	"errors"
	"tracker/internal/cache"
	"tracker/internal/db/models"
	repositories "tracker/internal/db/repo"
)

var ErrNotFound = errors.New("not found")

type PriceService struct {
	cache      *cache.RedisClient
	priceRepo  *repositories.PriceRepository
	fetcher    *PriceFetcher
	priceCache *PriceCache
}

func NewPriceService(
	cache *cache.RedisClient,
	priceRepo *repositories.PriceRepository,
	fetcher *PriceFetcher,
	priceCache *PriceCache,
) *PriceService {
	return &PriceService{
		cache:      cache,
		priceRepo:  priceRepo,
		fetcher:    fetcher,
		priceCache: priceCache,
	}
}

func (s *PriceService) GetCoins(ctx context.Context) ([]models.Coin, error) {
	coins, err := s.priceCache.GetCoins(ctx)
	if err == nil {
		return coins, nil
	}
	return nil, ErrNotFound
}

func (s *PriceService) GetCoin(ctx context.Context, id string) (*models.Coin, error) {
	coin := s.priceCache.GetCoinBySymbol(ctx, id)
	if coin != nil {
		return coin, nil
	}
	return nil, ErrNotFound
}

func (s *PriceService) GetPrices(ctx context.Context, symbols []string) ([]models.CoinPrice, error) {
	prices := []models.CoinPrice{}
	s.priceCache.AddPricesToWatch(ctx, symbols)
	for _, symbol := range symbols {
		price := s.priceCache.GetPriceBySymbol(ctx, symbol)
		if price != nil {
			prices = append(prices, *price)
		}
	}
	return prices, nil
}

func (s *PriceService) GetPrice(ctx context.Context, symbol string) (*models.CoinPrice, error) {
	s.priceCache.AddPricesToWatch(ctx, []string{symbol})
	price := s.priceCache.GetPriceBySymbol(ctx, symbol)
	if price != nil {
		return price, nil
	}
	return nil, ErrNotFound
}
