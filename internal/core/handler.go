package core

import (
	"context"
	"tracker/internal/cache"
	"tracker/internal/client"
)

type PriceService struct {
	priceClient *client.PriceClient // <- DEPENDENCY INJECTED
	cache       *cache.RedisClient  // <- DEPENDENCY INJECTED
}

func NewPriceService(priceClient *client.PriceClient, cache *cache.RedisClient) *PriceService {
	return &PriceService{
		priceClient: priceClient, // Store for later use
		cache:       cache,       // Store for later use
	}
}

func (s *PriceService) GetPrices(ctx context.Context, symbols []string) ([]client.PriceData, error) {
	// 1. Check cache using s.cache
	// 2. If miss, call s.priceClient.GetPrices()
	// 3. Cache result
	// 4. Return
	prices := []client.PriceData{}
	for _, symbol := range symbols {
		price := client.PriceData{}
		err := s.cache.GetJSON(ctx, symbol, price)
		if err != nil {
			prices = append(prices, price)
		}
	}
	return prices, nil
}
