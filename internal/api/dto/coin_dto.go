package dto

import (
	"tracker/internal/core/entity"
)

type CoinResponse struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

func ToCoinsResponse(coins []entity.Coin) []CoinResponse {
	result := make([]CoinResponse, len(coins))
	for i, coin := range coins {
		result[i] = ToCoinResponse(coin)
	}
	return result
}

func ToCoinResponse(coin entity.Coin) CoinResponse {
	return CoinResponse{
		Symbol:   coin.Symbol,
		Name:     coin.Name,
		ImageURL: coin.ImageURL,
	}
}
