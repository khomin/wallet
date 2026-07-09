package core

import (
	"context"
	"fmt"
	"tracker/internal/cache"
	"tracker/internal/core/entity"
	"tracker/internal/db/models"
	repositories "tracker/internal/db/repo"
)

type PriceService struct {
	cache     *cache.RedisClient
	priceRepo *repositories.PriceRepository
	fetcher   *PriceFetcher
}

func NewPriceService(
	cache *cache.RedisClient,
	priceRepo *repositories.PriceRepository,
	fetcher *PriceFetcher,
) *PriceService {
	return &PriceService{
		cache:     cache,
		priceRepo: priceRepo,
		fetcher:   fetcher,
	}
}

func (s *PriceService) GetCoinsSnapshot(ctx context.Context) ([]entity.Coin, error) {
	var cached []models.CoinSnapshot
	if err := s.cache.GetJSON(ctx, "coins:all", &cached); err != nil {
		return []entity.Coin{}, err
	}
	coins := []entity.Coin{}
	for _, i := range cached {
		coins = append(coins, entity.Coin{
			ID:       i.CoinID,
			Name:     i.Name,
			Symbol:   i.Symbol,
			ImageURL: i.ImageURL,
		})
	}
	return coins, nil
}

func (s *PriceService) GetCoinSnapshot(ctx context.Context, id string) (*entity.Coin, error) {
	cached := models.CoinSnapshot{}
	if err := s.cache.GetJSON(ctx, fmt.Sprintf("coins:%s", id), &cached); err != nil {
		return nil, err
	}
	return &entity.Coin{
		ID:       cached.CoinID,
		Name:     cached.Name,
		Symbol:   cached.Symbol,
		ImageURL: cached.ImageURL,
	}, nil
}

func (s *PriceService) GetPrices(ctx context.Context, symbols []string) ([]entity.Price, error) {
	prices := []entity.Price{}
	for _, symbol := range symbols {
		price := entity.Price{}
		err := s.cache.GetJSON(ctx, fmt.Sprintf("prices:%s", symbol), price)
		if err == nil {
			prices = append(prices, price)
		}
	}
	s.fetcher.setPricesToWatch(ctx, symbols)
	return prices, nil
}
