package models

import (
	"time"
)

type Coin struct {
	ID          string    `db:"id"`
	Symbol      string    `db:"symbol"` // "btc"
	Name        string    `db:"name"`   // "Bitcoin"
	ImageURL    string    `db:"image_url"`
	LastUpdated time.Time `db:"last_updated"`
	SnapshotAt  time.Time `db:"snapshot_at"` // When we captured this
}

type CoinPrice struct {
	ID                             string    `db:"id"`
	Name                           string    `db:"name"`
	Symbol                         string    `db:"symbol"`
	CurrentPrice                   float64   `db:"current_price"`
	Change_24h                     float64   `db:"change_24h"`
	MarketCap                      float64   `db:"market_cap"`
	TotalVolume                    float64   `db:"total_volume"`
	High_24h                       float64   `db:"high_24h"`
	Low_24h                        float64   `db:"low_24h"`
	PriceChange_24h                float64   `db:"price_change_24h"`
	PriceChangePercentage_24h      float64   `db:"price_change_percentage_24h"`
	MarketCapChange_24h            float64   `db:"market_cap_change_24h"`
	MarketCapChange_percentage_24h float64   `db:"market_cap_change_percentage_24h"`
	LastUpdated                    time.Time `db:"last_updated"`
}
