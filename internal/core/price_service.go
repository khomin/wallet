package core

import (
	"context"
	"fmt"
	"strings"
	"time"
	"tracker/internal/cache"
	"tracker/internal/client/alchemy"
	"tracker/internal/client/coingecko"
	"tracker/internal/core/entity"
	"tracker/internal/db/models"
	repositories "tracker/internal/db/repo"

	"github.com/sirupsen/logrus"
)

type PriceService struct {
	priceClient     *alchemy.AlchemyClient
	cache           *cache.RedisClient
	coingeckoClient *coingecko.CoinGeckoClient
	priceRepo       *repositories.PriceRepository
}

func NewPriceService(
	priceClient *alchemy.AlchemyClient,
	coingeckoClient *coingecko.CoinGeckoClient,
	cache *cache.RedisClient,
	priceRepo *repositories.PriceRepository,
) *PriceService {
	return &PriceService{
		priceClient:     priceClient,
		cache:           cache,
		coingeckoClient: coingeckoClient,
		priceRepo:       priceRepo,
	}
}

func (s *PriceService) GetCoins(ctx context.Context) ([]entity.Coin, error) {
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

func (s *PriceService) GetCoin(ctx context.Context, id string) (*entity.Coin, error) {
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
	log := logrus.WithField("PriceService", "GetPrices")
	prices := []entity.Price{}
	if len(symbols) == 0 {
		err := fmt.Errorf("empty symbols")
		log.WithError(err)
		return prices, err
	}
	symbolsUnique := make(map[string]string)
	for _, symbol := range symbols {
		symbolLower := strings.ToLower(symbol)
		if _, exists := symbolsUnique[symbolLower]; !exists {
			symbolsUnique[symbol] = symbolLower
		}
	}
	for symbol := range symbolsUnique {
		price := entity.Price{}
		err := s.cache.GetJSON(ctx, fmt.Sprintf("price:%s", symbol), price)
		if err == nil {
			prices = append(prices, price)
		}
		s.cache.Set(ctx, fmt.Sprintf("active-coin:%s", symbol), nil, 5*time.Minute)
	}
	return prices, nil
}
