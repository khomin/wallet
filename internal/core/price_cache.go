package core

import (
	"context"
	"fmt"
	"time"
	"tracker/internal/cache"
	"tracker/internal/core/entity"
)

type PriceCache struct {
	cache *cache.RedisClient
}

func NewPriceCache(cache *cache.RedisClient) *PriceCache {
	return &PriceCache{
		cache: cache,
	}
}

func (p *PriceCache) GetPriceBySymbol(ctx context.Context, symbol string) *entity.Price {
	price := entity.Price{}
	err := p.cache.GetJSON(ctx, fmt.Sprintf("prices:%s", symbol), price)
	if err == nil {
		return &price
	}
	return nil
}

func (p *PriceCache) GetCoins(ctx context.Context) ([]entity.Coin, error) {
	var coins []entity.Coin
	if err := p.cache.GetJSON(ctx, "coins:all", &coins); err != nil {
		return nil, err
	}
	// coins := []entity.Coin{}
	// for _, symbol := range symbols {
	// 	var coin entity.Coin
	// 	if err := p.cache.GetJSON(ctx, fmt.Sprintf("coins:%s", symbol), &coin); err != nil {
	// 		return coins
	// 	}
	// 	coins = append(coins, coin)
	// }
	return coins, nil
}

func (p *PriceCache) GetCoinsBySymbol(ctx context.Context, symbols []string) []entity.Coin {
	coins := []entity.Coin{}
	for _, symbol := range symbols {
		var coin entity.Coin
		if err := p.cache.GetJSON(ctx, fmt.Sprintf("coins:%s", symbol), &coin); err != nil {
			return coins
		}
		coins = append(coins, coin)
	}
	return coins
}

func (p *PriceCache) GetCoinBySymbol(ctx context.Context, symbol string) *entity.Coin {
	var coin entity.Coin
	if err := p.cache.GetJSON(ctx, fmt.Sprintf("coins:%s", symbol), &coin); err != nil {
		return nil
	}
	return &coin
}

func (p *PriceCache) SetCoins(ctx context.Context, snapshots []entity.Coin) error {
	if err := p.cache.SetJSON(ctx, "coins:all", snapshots, 1*time.Hour); err != nil {
		return err
	}
	return nil
}

// if err := s.cache.GetJSON(ctx, fmt.Sprintf("coins:%s", id), &cached); err != nil {
// 	return nil, err
// }
// return &entity.Coin{
// 	ID:       cached.CoinID,
// 	Name:     cached.Name,
// 	Symbol:   cached.Symbol,
// 	ImageURL: cached.ImageURL,
// }, nil
