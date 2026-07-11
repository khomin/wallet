package dto

import (
	"tracker/internal/db/models"
)

type PriceResponse struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	CurrentPrice float64 `json:"current_price"`
}

func ToPricesResponse(prices []models.Price) []PriceResponse {
	result := make([]PriceResponse, 0)
	for i, v := range prices {
		result[i] = ToPriceResponse(v)
	}
	return result
}

func ToPriceResponse(price models.Price) PriceResponse {
	return PriceResponse{
		Name:         price.Name,
		Symbol:       price.Symbol,
		CurrentPrice: price.PriceUSD,
	}
}
