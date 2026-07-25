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

func ToAssetsResponse(in []entity.Asset) AssetsResponse {
	out := []Coin{}
	for _, v := range in {
		out = append(out, Coin{
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
