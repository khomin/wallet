package dto

import "tracker/internal/core/domain"

type CoinsResponse struct {
	Total int                  `json:"total"`
	Coins []domain.TokenSimple `json:"coins"`
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
