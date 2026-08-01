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
	GetCoinSnapshot(ctx context.Context) ([]domain.TokenSimple, error)
	SetCoinSnapshot(ctx context.Context, in []domain.TokenSimple) error
	GetPriceSnapshot(ctx context.Context) ([]domain.TokenPrice, error)
	SetPriceSnapshot(ctx context.Context, in []domain.TokenPrice) error
}

type AssetService struct {
	cache         *cache.RedisClient
	priceRepo     PriceRepository
	fetcher       *PriceFetcher
	priceCache    *PriceCache
	tokenRegistry *Registry
}

func NewPriceService(
	cache *cache.RedisClient,
	priceRepo PriceRepository,
	fetcher *PriceFetcher,
	priceCache *PriceCache,
	tokenRegistry *Registry,
) *AssetService {
	return &AssetService{
		cache:         cache,
		priceRepo:     priceRepo,
		fetcher:       fetcher,
		priceCache:    priceCache,
		tokenRegistry: tokenRegistry,
	}
}

func (s *AssetService) GetCoins(ctx context.Context) ([]domain.TokenSimple, error) {
	coins, err := s.priceCache.GetCoins(ctx)
	if err != nil {
		return nil, err
	}
	tokens := []domain.TokenSimple{}
	for _, coin := range coins {
		token, found := s.tokenRegistry.GetBySymbol(coin.Symbol)
		if found {
			tokens = append(tokens, domain.TokenSimple{
				Symbol:   token.Symbol,
				Name:     token.Name,
				ImageURL: coin.ImageURL,
			})
		}
	}
	return tokens, nil
}

func (s *AssetService) GetCoin(ctx context.Context, id string) (*domain.TokenSimple, error) {
	coin := s.priceCache.GetCoinBySymbol(ctx, id)
	if coin == nil {
		return nil, ErrPriceNotFound
	}
	token, found := s.tokenRegistry.GetBySymbol(coin.Symbol)
	if found {
		return &domain.TokenSimple{
			Symbol:   token.Symbol,
			Name:     token.Name,
			ImageURL: coin.ImageURL,
		}, nil
	}
	return nil, ErrPriceNotFound
}

func (s *AssetService) SearchCoins(ctx context.Context, text string) ([]domain.TokenSimple, error) {
	out := []domain.TokenSimple{}
	tokens := s.tokenRegistry.GetByQuery(text)
	for _, i := range tokens {
		coin := s.priceCache.GetCoinBySymbol(ctx, i.Symbol)
		out = append(out, domain.TokenSimple{
			Symbol:   i.Symbol,
			Name:     i.Name,
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
	// add prices to watch
	s.priceCache.AddPricesToWatch(ctx, symbols)
	return prices, nil
}

func (s *AssetService) GetPrice(ctx context.Context, symbol string) (domain.TokenPrice, error) {
	price := s.priceCache.GetPriceBySymbol(ctx, symbol)
	if price != nil {
		s.priceCache.AddPricesToWatch(ctx, []string{symbol})
		return *price, nil
	}
	return domain.TokenPrice{}, ErrPriceNotFound
}
