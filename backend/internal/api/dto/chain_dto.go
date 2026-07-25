package dto

import (
	"tracker/internal/core/entity"
)

// type Chain struct {
// 	Symbol   string `json:"symbol"`
// 	Name     string `json:"name"`
// 	ImageURL string `json:"image_url"`
// }

type AssetsResponse struct {
	Total  int    `json:"total"`
	Assets []Coin `json:"assets"`
}

func ToAssetsResponse(chains []entity.Token) AssetsResponse {
	out := []Coin{}
	for _, chain := range chains {
		out = append(out, Coin{
			Symbol:   chain.Symbol,
			Name:     chain.Name,
			ImageURL: chain.LogoURL,
		})
	}
	return AssetsResponse{
		Total:  len(out),
		Assets: out,
	}
}

// func ToAssetssResponse(chains []entity.Token, coins []models.Coin) ChainsResponse {
// 	out := []Chain{}
// 	findCoin := func(symbol string) *models.Coin {
// 		for _, i := range coins {
// 			if strings.EqualFold(symbol, i.Symbol) {
// 				return &i
// 			}
// 		}
// 		return nil
// 	}
// 	for _, chain := range chains {
// 		coin := findCoin(chain.Symbol)
// 		if coin == nil || coin.ImageURL == "" {
// 			continue
// 		}
// 		out = append(out, Chain{
// 			Symbol:   chain.Symbol,
// 			Name:     chain.Name,
// 			ImageURL: coin.ImageURL,
// 		})
// 	}
// 	return ChainsResponse{
// 		Total:  len(out),
// 		Chains: out,
// 	}
// }
