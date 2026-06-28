package models

import "time"

// CachedPrice - stored in Redis for fast access
type CachedPrice struct {
	Symbol      string    `json:"symbol"`
	PriceUSD    float64   `json:"price_usd"`
	Change24h   float64   `json:"change_24h"`
	MarketCap   float64   `json:"market_cap"`
	Volume24h   float64   `json:"volume_24h"`
	LastUpdated time.Time `json:"last_updated"`
}

// CachedPriceList - stores multiple prices in one Redis key
type CachedPriceList struct {
	Prices    []CachedPrice `json:"prices"`
	UpdatedAt time.Time     `json:"updated_at"`
}
