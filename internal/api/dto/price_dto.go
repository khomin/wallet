package dto

import (
	"time"
	"tracker/internal/core/entity"
)

type PriceResponse struct {
	Symbol                    string      `json:"symbol"`
	Name                      string      `json:"name"`
	Image                     string      `json:"image"`
	CurrentPrice              float64     `json:"current_price"`
	MarketCap                 float64     `json:"market_cap"`
	MarketCapRank             int         `json:"market_cap_rank"`
	FullyDilutedValuation     float64     `json:"fully_diluted_valuation"`
	TotalVolume               float64     `json:"total_volume"`
	High24h                   float64     `json:"high_24h"`
	Low24h                    float64     `json:"low_24h"`
	PriceChange24h            float64     `json:"price_change_24h"`
	PriceChangePercent24h     float64     `json:"price_change_percentage_24h"`
	MarketCapChange24h        float64     `json:"market_cap_change_24h"`
	MarketCapChangePercent24h float64     `json:"market_cap_change_percentage_24h"`
	CirculatingSupply         float64     `json:"circulating_supply"`
	TotalSupply               *float64    `json:"total_supply"`
	MaxSupply                 *float64    `json:"max_supply"`
	ATH                       float64     `json:"ath"`
	ATHChangePercent          float64     `json:"ath_change_percentage"`
	ATHDate                   time.Time   `json:"ath_date"`
	ATL                       float64     `json:"atl"`
	ATLChangePercent          float64     `json:"atl_change_percentage"`
	ATLDate                   time.Time   `json:"atl_date"`
	ROI                       interface{} `json:"roi"`
	LastUpdated               time.Time   `json:"last_updated"`
}

func ToPriceResponse(prices []entity.Price) []PriceResponse {
	result := make([]PriceResponse, 0)
	for i, coin := range prices {
		result[i] = PriceResponse{
			Symbol: coin.Symbol,
			Name:   coin.Symbol,
		}
	}
	return result
}
