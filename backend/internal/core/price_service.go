package core

import (
	"context"
	"errors"
	"slices"
	"tracker/internal/cache"
	"tracker/internal/core/domain"
)

var ErrPriceNotFound = errors.New("not found")

type PriceRepository interface {
	GetCoinSnapshot(ctx context.Context) ([]domain.TokenID, error)
	SetCoinSnapshot(ctx context.Context, in []domain.TokenID) error
	GetPriceSnapshot(ctx context.Context) ([]domain.TokenPrice, error)
	SetPriceSnapshot(ctx context.Context, in []domain.TokenPrice) error
}

type AssetService struct {
	cache         *cache.RedisClient
	priceRepo     PriceRepository
	fetcher       *PriceFetcher
	priceCache    PriceCache
	tokenRegistry *TokenRegistry
}

func NewPriceService(
	cache *cache.RedisClient,
	priceRepo PriceRepository,
	fetcher *PriceFetcher,
	priceCache PriceCache,
	tokenRegistry *TokenRegistry,
) *AssetService {
	return &AssetService{
		cache:         cache,
		priceRepo:     priceRepo,
		fetcher:       fetcher,
		priceCache:    priceCache,
		tokenRegistry: tokenRegistry,
	}
}

func (s *AssetService) GetCoins(ctx context.Context) ([]domain.TokenWithURL, error) {
	coins, err := s.priceCache.GetCoins(ctx)
	if err != nil {
		return nil, err
	}
	tokens := []domain.TokenWithURL{}
	for _, coin := range coins {
		token, found := s.tokenRegistry.GetBySymbol(coin.Symbol)
		if found {
			tokens = append(tokens, domain.TokenWithURL{
				TokenRaw: token,
				ImageURL: coin.ImageURL,
			})
		}
	}
	return tokens, nil
}

func (s *AssetService) GetCoin(ctx context.Context, id string) (*domain.TokenWithURL, error) {
	coin := s.priceCache.GetCoinBySymbol(ctx, id)
	if coin == nil {
		return nil, ErrPriceNotFound
	}
	token, found := s.tokenRegistry.GetBySymbol(coin.Symbol)
	if found {
		return &domain.TokenWithURL{
			TokenRaw: token,
			ImageURL: coin.ImageURL,
		}, nil
	}
	return nil, ErrPriceNotFound
}

func (s *AssetService) SearchCoins(ctx context.Context, text string) ([]domain.TokenWithURL, error) {
	out := []domain.TokenWithURL{}
	tokens := s.tokenRegistry.GetByQuery(text)
	for _, token := range tokens {
		coin := s.priceCache.GetCoinBySymbol(ctx, token.Symbol)
		out = append(out, domain.TokenWithURL{
			TokenRaw: token,
			ImageURL: coin.ImageURL,
		})
	}
	return out, nil
}

func (s *AssetService) GetPrices(ctx context.Context, symbols []string) ([]domain.TokenPrice, error) {
	// keep only supported symbols
	symbols = slices.DeleteFunc(symbols, func(symbol string) bool {
		_, found := s.tokenRegistry.GetBySymbol(symbol)
		return !found
	})
	// get prices
	prices := []domain.TokenPrice{}
	for _, symbol := range symbols {
		price := s.priceCache.GetPriceBySymbol(ctx, symbol)
		if price != nil {
			prices = append(prices, *price)
		}
	}
	return prices, nil
}

func (s *AssetService) GetPrice(ctx context.Context, symbol string) (domain.TokenPrice, error) {
	price := s.priceCache.GetPriceBySymbol(ctx, symbol)
	if price != nil {
		return *price, nil
	}
	return domain.TokenPrice{}, ErrPriceNotFound
}
