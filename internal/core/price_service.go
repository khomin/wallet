package core

import (
	"context"
	"errors"
	"fmt"
	"tracker/internal/cache"
	"tracker/internal/db/models"
	repositories "tracker/internal/db/repo"
)

var ErrPriceNotFound = errors.New("price not found")

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
	if err != nil {
		return nil, err
	}
	return coins, nil
}

func (s *PriceService) GetCoinSnapshot(ctx context.Context, id string) (*models.Coin, error) {
	if coin := s.priceCache.GetCoinBySymbol(ctx, id); coin != nil {
		return coin, nil
	}
	return nil, fmt.Errorf("not found")
}

func (s *PriceService) GetPrices(ctx context.Context, symbols []string) ([]models.Price, error) {
	prices := []models.Price{}
	for _, symbol := range symbols {
		price := s.priceCache.GetPriceBySymbol(ctx, symbol)
		if price != nil {
			prices = append(prices, *price)
		}
	}
	s.fetcher.addPricesToWatch(ctx, symbols)
	return prices, nil
}

func (s *PriceService) GetPrice(ctx context.Context, symbol string) (*models.Price, error) {
	s.fetcher.addPricesToWatch(ctx, []string{symbol})
	price := s.priceCache.GetPriceBySymbol(ctx, symbol)
	if price != nil {
		return price, nil
	}
	return nil, ErrPriceNotFound
}
