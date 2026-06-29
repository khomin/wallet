package core

import (
	"context"
	"tracker/internal/cache"
	"tracker/internal/client"
)

type PriceService struct {
	priceClient *client.AlchemyClient
	cache       *cache.RedisClient
}

func NewPriceService(priceClient *client.AlchemyClient, cache *cache.RedisClient) *PriceService {
	return &PriceService{
		priceClient: priceClient,
		cache:       cache,
	}
}

// prices:all

func (s *PriceService) GetPricesAll(ctx context.Context) ([]client.PriceData, error) {
	return s.GetPrices(ctx, nil)
}

func (s *PriceService) GetPrices(ctx context.Context, symbols []string) ([]client.PriceData, error) {
	// 1. Check cache using s.cache
	// 2. If miss, call s.priceClient.GetPrices()
	// 3. Cache result
	// 4. Return
	s.priceClient.GetPrices()
	prices := []client.PriceData{}
	if symbols == nil {
		var coins []client.CoinGeckoCoin
		if err := s.cache.GetJSON(ctx, "prices:all", coins); err == nil {
			return coins, nil
		}
	} else {
		for _, symbol := range symbols {
			price := client.PriceData{}
			err := s.cache.GetJSON(ctx, symbol, price)
			if err != nil {
				prices = append(prices, price)
			} else {
				s.priceClient.GetPrices()
			}
		}
	}
	return prices, nil
}
