package entity

import "time"

type Price struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	PriceUSD    float64   `json:"price_usd"`
	Change24h   float64   `json:"change_24h"`
	LastUpdated time.Time `json:"last_updated"`
}
