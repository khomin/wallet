package models

import (
	"time"

	"github.com/google/uuid"
)

type Coin struct {
	ID          uuid.UUID `db:"id" json:"id"`
	CoinID      string    `db:"coin_id" json:"coin_id"` // "bitcoin"
	Symbol      string    `db:"symbol" json:"symbol"`   // "btc"
	Name        string    `db:"name" json:"name"`       // "Bitcoin"
	ImageURL    string    `db:"image_url" json:"image_url"`
	LastUpdated time.Time `db:"last_updated" json:"last_updated"`
	SnapshotAt  time.Time `db:"snapshot_at" json:"snapshot_at"` // When we captured this
}

type Price struct {
	ID                             uuid.UUID `db:"id" json:"id"`
	CoinID                         string    `db:"coin_id" json:"coin_id"` // "bitcoin"
	Symbol                         string    `db:"id" json:"symbol"`
	Name                           string    `db:"name" json:"name"` // "Bitcoin"
	CurrentPrice                   float64   `db:"current_price" json:"current_price"`
	Change_24h                     float64   `db:"change_24h" json:"change_24h"`
	MarketCap                      float64   `db:"market_cap" json:"market_cap"`
	TotalVolume                    float64   `db:"total_volume" json:"total_volume"`
	High_24h                       float64   `db:"high_24h" json:"high_24h"`
	Low_24h                        float64   `db:"low_24h" json:"low_24h"`
	PriceChange_24h                float64   `db:"price_change_24h" json:"price_change_24h"`
	PriceChangePercentage_24h      float64   `db:"price_change_percentage_24h" json:"price_change_percentage_24h"`
	MarketCapChange_24h            float64   `db:"market_cap_change_24h" json:"market_cap_change_24h"`
	MarketCapChange_percentage_24h float64   `db:"market_cap_change_percentage_24h" json:"market_cap_change_percentage_24h"`
	LastUpdated                    time.Time `db:"last_updated" json:"last_updated"`
}
