package core

import (
	"context"
	"errors"
	"slices"
	"tracker/internal/cache"
	"tracker/internal/core/entity"
	"tracker/internal/db/models"
)

var ErrPriceNotFound = errors.New("not found")

type PriceRepository interface {
	GetCoinSnapshot(ctx context.Context) ([]models.Coin, error)
	SetCoinSnapshot(ctx context.Context, snapshots []models.Coin) error
	GetPriceSnapshot(ctx context.Context) ([]models.CoinPrice, error)
	SetPriceSnapshot(ctx context.Context, snapshots []models.CoinPrice) error
}

type AssetService struct {
	cache         *cache.RedisClient
	priceRepo     PriceRepository
	fetcher       *PriceFetcher
	priceCache    *PriceCache
	tokenRegistry *TokenRegistry
}

func NewPriceService(
	cache *cache.RedisClient,
	priceRepo PriceRepository,
	fetcher *PriceFetcher,
	priceCache *PriceCache,
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

func (s *AssetService) GetCoins(ctx context.Context) ([]entity.Token, error) {
	coins, err := s.priceCache.GetCoins(ctx)
	if err != nil {
		return nil, err
	}
	tokens := []entity.Token{}
	for _, coin := range coins {
		token, found := s.tokenRegistry.GetBySymbol(coin.Symbol)
		if found {
			token.AddImageURL(coin.ImageURL)
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
}

func (s *AssetService) GetCoin(ctx context.Context, id string) (entity.Token, error) {
	coin := s.priceCache.GetCoinBySymbol(ctx, id)
	if coin == nil {
		return entity.Token{}, ErrPriceNotFound
	}
	token, found := s.tokenRegistry.GetBySymbol(coin.Symbol)
	if found {
		token.AddImageURL(coin.ImageURL)
		return token, nil
	}
	return entity.Token{}, ErrPriceNotFound
}

func (s *AssetService) SearchCoins(ctx context.Context, text string) ([]entity.Token, error) {
	tokens := []entity.Token{}
	assets := s.tokenRegistry.GetAssetsByText(text)
	for _, asset := range assets {
		for _, token := range asset.Tokens {
			symbol := token.Symbol
			if symbol == "" {
				symbol = asset.Symbol
			}
			token, found := s.tokenRegistry.GetBySymbol(symbol)
			if found {
				token.AddImageURL(token.ImageURL)
				tokens = append(tokens, token)
			}
		}
	}

	// coins, err := s.priceCache.GetCoins(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	return tokens, nil
}

func (s *AssetService) GetPrices(ctx context.Context, symbols []string) ([]entity.TokenPrice, error) {
	// keep only supported symbols
	symbols = slices.DeleteFunc(symbols, func(symbol string) bool {
		_, found := s.tokenRegistry.GetBySymbol(symbol)
		return !found
	})
	// get prices
	prices := []entity.TokenPrice{}
	for _, symbol := range symbols {
		price := s.priceCache.GetPriceBySymbol(ctx, symbol)
		if price != nil {
			prices = append(prices, dbPriceToEntity(price))
		}
	}
	// add prices to watch
	s.priceCache.AddPricesToWatch(ctx, symbols)
	return prices, nil
}

func (s *AssetService) GetPrice(ctx context.Context, symbol string) (entity.TokenPrice, error) {
	price := s.priceCache.GetPriceBySymbol(ctx, symbol)
	if price != nil {
		s.priceCache.AddPricesToWatch(ctx, []string{symbol})
		return dbPriceToEntity(price), nil
	}
	return entity.TokenPrice{}, ErrPriceNotFound
}

func dbCoinsToEntity(in []models.Coin) []entity.Token {
	out := []entity.Token{}
	for _, i := range in {
		out = append(out, entity.Token{
			Chain:    i.Symbol,
			Symbol:   i.Symbol,
			Name:     i.Name,
			Address:  i.Name,
			Decimals: 0, IsNative: false,
			ImageURL: i.ImageURL,
		})
	}
	return out
}

func dbPriceToEntity(in *models.CoinPrice) entity.TokenPrice {
	return entity.TokenPrice{
		ID:                             in.ID,
		CoinID:                         in.CoinID,
		Name:                           in.Name,
		Symbol:                         in.Symbol,
		CurrentPrice:                   in.CurrentPrice,
		Change_24h:                     in.Change_24h,
		MarketCap:                      in.MarketCap,
		TotalVolume:                    in.TotalVolume,
		High_24h:                       in.High_24h,
		Low_24h:                        in.Low_24h,
		PriceChange_24h:                in.PriceChange_24h,
		PriceChangePercentage_24h:      in.PriceChangePercentage_24h,
		MarketCapChange_24h:            in.MarketCapChange_24h,
		MarketCapChange_percentage_24h: in.MarketCapChange_percentage_24h,
		LastUpdated:                    in.LastUpdated,
	}
}
