package dto

import (
	"time"
	"tracker/internal/db/models"
)

type PriceResponse struct {
	Symbol                         string    `json:"symbol"`
	Name                           string    `json:"name"`
	CurrentPrice                   float64   `json:"current_price"`
	MarketCap                      float64   `json:"market_cap"`
	TotalVolume                    float64   `json:"total_volume"`
	High_24h                       float64   `json:"high_24h"`
	Low_24h                        float64   `json:"low_24h"`
	PriceChange_24h                float64   `json:"price_change_24h"`
	PriceChangePercentage_24h      float64   `json:"price_change_percentage_24h"`
	MarketCapChange_24h            float64   `json:"market_cap_change_24h"`
	MarketCapChange_percentage_24h float64   `json:"market_cap_change_percentage_24h"`
	LastUpdated                    time.Time `json:"last_updated"`
}

func ToPricesResponse(prices []models.CoinPrice) []PriceResponse {
	result := make([]PriceResponse, len(prices))
	for i, v := range prices {
		result[i] = ToPriceResponse(v)
	}
	return result
}

func ToPriceResponse(price models.CoinPrice) PriceResponse {
	return PriceResponse{
		Symbol:                         price.Symbol,
		Name:                           price.Name,
		CurrentPrice:                   price.CurrentPrice,
		MarketCap:                      price.MarketCap,
		TotalVolume:                    price.TotalVolume,
		High_24h:                       price.High_24h,
		Low_24h:                        price.Low_24h,
		PriceChange_24h:                price.PriceChange_24h,
		PriceChangePercentage_24h:      price.PriceChangePercentage_24h,
		MarketCapChange_24h:            price.MarketCapChange_24h,
		MarketCapChange_percentage_24h: price.MarketCapChange_percentage_24h,
		LastUpdated:                    price.LastUpdated,
	}
}
