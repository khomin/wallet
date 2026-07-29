package dto

import (
	"tracker/internal/core/entity"
)

type CoinsResponse struct {
	Total int     `json:"total"`
	Coins []Token `json:"coins"`
}

type Token struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

type AssetsResponse struct {
	Total  int     `json:"total"`
	Assets []Token `json:"assets"`
}

type SearchCoins struct {
	Text string `json:"text"`
}

func ToCoinsResponse(coins []entity.Token) CoinsResponse {
	coins_ := make([]Token, len(coins))
	for i, coin := range coins {
		coins_[i] = ToCoinResponse(coin)
	}
	return CoinsResponse{
		Total: len(coins_),
		Coins: coins_,
	}
}

func ToCoinResponse(coin entity.Token) Token {
	return Token{
		Symbol:   coin.Symbol,
		Name:     coin.Name,
		ImageURL: coin.ImageURL,
	}
}

func ToAssetsResponse(in []entity.Asset) AssetsResponse {
	out := []Token{}
	for _, v := range in {
		out = append(out, Token{
			Symbol: v.Symbol,
			Name:   v.Name,
			// ImageURL: v. LogoURL,
		})
	}
	return AssetsResponse{
		Total:  len(out),
		Assets: out,
	}
}
