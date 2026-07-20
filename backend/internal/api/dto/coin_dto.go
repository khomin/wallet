package dto

import "tracker/internal/db/models"

type CoinsResponse struct {
	Total int            `json:"total"`
	Coins []CoinResponse `json:"coins"`
}

type CoinResponse struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

func ToCoinsResponse(coins []models.Coin) CoinsResponse {
	coins_ := make([]CoinResponse, len(coins))
	for i, coin := range coins {
		coins_[i] = ToCoinResponse(&coin)
	}
	return CoinsResponse{
		Total: len(coins_),
		Coins: coins_,
	}
}

func ToCoinResponse(coin *models.Coin) CoinResponse {
	return CoinResponse{
		Symbol:   coin.Symbol,
		Name:     coin.Name,
		ImageURL: coin.ImageURL,
	}
}
