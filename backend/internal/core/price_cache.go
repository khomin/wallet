package core

import (
	"context"
	"fmt"
	"strings"
	"time"
	"tracker/internal/cache"
	"tracker/internal/core/domain"
)

type PriceCache interface {
	GetPriceBySymbol(ctx context.Context, symbol string) *domain.TokenPrice
	SetPrices(ctx context.Context, prices []domain.TokenPrice) error
	SetPrice(ctx context.Context, symbol string, price domain.TokenPrice) error
	GetCoins(ctx context.Context) ([]domain.TokenID, error)
	GetCoinsBySymbol(ctx context.Context, symbols []string) ([]domain.TokenID, error)
	GetCoinBySymbol(ctx context.Context, symbol string) *domain.TokenID
	SetCoins(ctx context.Context, coins []domain.TokenID) error

	GetBalanceNative(ctx context.Context, chain, address string) (float64, error)
	SetBalanceNative(ctx context.Context, chain, address string, balance float64) error
	GetBalanceToken(ctx context.Context, chain, address, tokenSymbol string) (float64, error)
	SetBalanceToken(ctx context.Context, chain, address, tokenSymbol string, balance float64) error
}

type noOpCache struct{}

func NewNoOpCache() PriceCache {
	return &noOpCache{}
}

func (p *noOpCache) GetPriceBySymbol(ctx context.Context, symbol string) *domain.TokenPrice {
	return nil
}
func (p *noOpCache) SetPrices(ctx context.Context, prices []domain.TokenPrice) error {
	return nil
}
func (p *noOpCache) SetPrice(ctx context.Context, symbol string, price domain.TokenPrice) error {
	return nil
}
func (p *noOpCache) GetCoins(ctx context.Context) ([]domain.TokenID, error) {
	return nil, nil
}
func (p *noOpCache) GetCoinsBySymbol(ctx context.Context, symbols []string) ([]domain.TokenID, error) {
	return nil, nil
}
func (p *noOpCache) GetCoinBySymbol(ctx context.Context, symbol string) *domain.TokenID {
	return nil
}
func (p *noOpCache) SetCoins(ctx context.Context, coins []domain.TokenID) error {
	return nil
}
func (p *noOpCache) AddPricesToWatch(ctx context.Context, symbols []string) error {
	return nil
}
func (p *noOpCache) GetPricesToWatch(ctx context.Context) []string {
	return nil
}
func (p *noOpCache) GetBalanceNative(ctx context.Context, chain, address string) (float64, error) {
	return 0, fmt.Errorf("no operational")
}
func (p *noOpCache) SetBalanceNative(ctx context.Context, chain, address string, balance float64) error {
	return nil
}
func (p *noOpCache) GetBalanceToken(ctx context.Context, chain, address, tokenSymbol string) (float64, error) {
	return 0, fmt.Errorf("no operational")
}
func (p *noOpCache) SetBalanceToken(ctx context.Context, chain, address, tokenSymbol string, balance float64) error {
	return nil
}

type priceCache struct {
	cache *cache.RedisClient
}

func NewPriceCache(cache *cache.RedisClient) *priceCache {
	return &priceCache{
		cache: cache,
	}
}

func (p *priceCache) GetPriceBySymbol(ctx context.Context, symbol string) *domain.TokenPrice {
	price := domain.TokenPrice{}
	err := p.cache.GetJSON(ctx, fmt.Sprintf("prices:%s", strings.ToUpper(symbol)), &price)
	if err == nil {
		return &price
	}
	return nil
}

func (p *priceCache) SetPrices(ctx context.Context, prices []domain.TokenPrice) error {
	for _, price := range prices {
		if err := p.cache.SetJSON(ctx, fmt.Sprintf("prices:%s", strings.ToUpper(price.Symbol)), price, 1*time.Hour); err != nil {
			return err
		}
	}
	return nil
}

func (p *priceCache) SetPrice(ctx context.Context, symbol string, price domain.TokenPrice) error {
	return p.cache.SetJSON(ctx, fmt.Sprintf("prices:%s", strings.ToUpper(symbol)), price, 1*time.Hour)
}

func (p *priceCache) GetCoins(ctx context.Context) ([]domain.TokenID, error) {
	var coins []domain.TokenID
	if err := p.cache.GetJSON(ctx, "coins:list", &coins); err != nil {
		return nil, err
	}
	return coins, nil
}

func (p *priceCache) GetCoinsBySymbol(ctx context.Context, symbols []string) ([]domain.TokenID, error) {
	coins := []domain.TokenID{}
	for _, symbol := range symbols {
		var coin domain.TokenID
		if err := p.cache.GetJSON(ctx, fmt.Sprintf("coins:%s", strings.ToUpper(symbol)), &coin); err != nil {
			return nil, err
		}
		coins = append(coins, coin)
	}
	return coins, nil
}

func (p *priceCache) GetCoinBySymbol(ctx context.Context, symbol string) *domain.TokenID {
	var coin domain.TokenID
	if err := p.cache.GetJSON(ctx, fmt.Sprintf("coins:%s", strings.ToUpper(symbol)), &coin); err != nil {
		return nil
	}
	return &coin
}

func (p *priceCache) SetCoins(ctx context.Context, coins []domain.TokenID) error {
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

func (p *priceCache) GetBalanceNative(ctx context.Context, chain, address string) (float64, error) {
	var v float64
	cacheKey := fmt.Sprintf("balance-n:%s:%s", chain, address)
	if err := p.cache.GetJSON(ctx, cacheKey, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (p *priceCache) SetBalanceNative(ctx context.Context, chain, address string, balance float64) error {
	cacheKey := fmt.Sprintf("balance-n:%s:%s", chain, address)
	if err := p.cache.SetJSON(ctx, cacheKey, balance, 1*time.Minute); err != nil {
		return err
	}
	return nil
}

func (p *priceCache) GetBalanceToken(ctx context.Context, chain, address, tokenSymbol string) (float64, error) {
	var v float64
	cacheKey := fmt.Sprintf("balance-t:%s:%s:%s", chain, address, tokenSymbol)
	if err := p.cache.GetJSON(ctx, cacheKey, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (p *priceCache) SetBalanceToken(ctx context.Context, chain, address, tokenSymbol string, balance float64) error {
	cacheKey := fmt.Sprintf("balance-t:%s:%s:%s", chain, address, tokenSymbol)
	if err := p.cache.SetJSON(ctx, cacheKey, balance, 1*time.Minute); err != nil {
		return err
	}
	return nil
}
