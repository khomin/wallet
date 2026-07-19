package core

import (
	"context"
	"fmt"
	"strings"
	"time"
	"tracker/internal/cache"
	"tracker/internal/db/models"
)

type PriceCache struct {
	cache *cache.RedisClient
}

func NewPriceCache(cache *cache.RedisClient) *PriceCache {
	return &PriceCache{
		cache: cache,
	}
}

func (p *PriceCache) GetPriceBySymbol(ctx context.Context, symbol string) *models.CoinPrice {
	price := models.CoinPrice{}
	err := p.cache.GetJSON(ctx, fmt.Sprintf("prices:%s", strings.ToUpper(symbol)), &price)
	if err == nil {
		return &price
	}
	return nil
}

func (p *PriceCache) SetPrices(ctx context.Context, prices []models.CoinPrice) error {
	for _, price := range prices {
		if err := p.cache.SetJSON(ctx, fmt.Sprintf("prices:%s", strings.ToUpper(price.Symbol)), price, 60*time.Second); err != nil {
			return err
		}
	}
	return nil
}

func (p *PriceCache) SetPrice(ctx context.Context, symbol string, price models.CoinPrice) error {
	return p.cache.SetJSON(ctx, fmt.Sprintf("prices:%s", strings.ToUpper(symbol)), price, 60*time.Second)
}

func (p *PriceCache) GetCoins(ctx context.Context) ([]models.Coin, error) {
	var coins []models.Coin
	if err := p.cache.GetJSON(ctx, "coins:list", &coins); err != nil {
		return nil, err
	}
	return coins, nil
}

func (p *PriceCache) GetCoinsBySymbol(ctx context.Context, symbols []string) ([]models.Coin, error) {
	coins := []models.Coin{}
	for _, symbol := range symbols {
		var coin models.Coin
		if err := p.cache.GetJSON(ctx, fmt.Sprintf("coins:%s", strings.ToUpper(symbol)), &coin); err != nil {
			return nil, err
		}
		coins = append(coins, coin)
	}
	return coins, nil
}

func (p *PriceCache) GetCoinBySymbol(ctx context.Context, symbol string) *models.Coin {
	var coin models.Coin
	if err := p.cache.GetJSON(ctx, fmt.Sprintf("coins:%s", strings.ToUpper(symbol)), &coin); err != nil {
		return nil
	}
	return &coin
}

func (p *PriceCache) SetCoins(ctx context.Context, coins []models.Coin) error {
	if err := p.cache.SetJSON(ctx, "coins:list", coins, 1*time.Hour); err != nil {
		return err
	}
	for _, i := range coins {
		if err := p.cache.SetJSON(ctx, fmt.Sprintf("coins:%s", strings.ToUpper(i.Symbol)), i, 1*time.Hour); err != nil {
			return err
		}
	}
	return nil
}

func (p *PriceCache) AddPricesToWatch(ctx context.Context, symbols []string) error {
	for _, symbol := range symbols {
		if err := p.cache.Set(ctx, fmt.Sprintf("prices-to-watch:%s", strings.ToUpper(symbol)), symbol, 5*time.Minute); err != nil {
			return err
		}
	}
	return nil
}

func (p *PriceCache) GetPricesToWatch(ctx context.Context) []string {
	prices := []string{}
	found, err := p.cache.Scan(ctx, "prices-to-watch:*")
	if err != nil {
		return prices
	}
	for _, foundPrice := range found {
		prices = append(prices, foundPrice.(string))
	}
	return prices
}
