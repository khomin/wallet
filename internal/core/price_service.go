package core

import (
	"context"
	"fmt"
	"tracker/internal/cache"
	"tracker/internal/core/entity"
	repositories "tracker/internal/db/repo"
)

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

func (s *PriceService) GetCoins(ctx context.Context) ([]entity.Coin, error) {
	coins, err := s.priceCache.GetCoins(ctx)
	if err != nil {
		return nil, err
	}
	return coins, nil
}

func (s *PriceService) GetCoinSnapshot(ctx context.Context, id string) (*entity.Coin, error) {
	if coin := s.priceCache.GetCoinBySymbol(ctx, id); coin != nil {
		return coin, nil
	}
	return nil, fmt.Errorf("not found")
}

func (s *PriceService) GetPrices(ctx context.Context, symbols []string) ([]entity.Price, error) {
	prices := []entity.Price{}
	for _, symbol := range symbols {
		price := s.priceCache.GetPriceBySymbol(ctx, symbol)
		if price != nil {
			prices = append(prices, *price)
		}
	}
	s.fetcher.setPricesToWatch(ctx, symbols)
	return prices, nil
}
