package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CoinSnapshot - stores daily/periodic snapshots of coin data
type CoinSnapshot struct {
	ID     uuid.UUID `db:"id" json:"id"`
	CoinID string    `db:"coin_id" json:"coin_id"` // "bitcoin"
	Symbol string    `db:"symbol" json:"symbol"`   // "btc"
	Name   string    `db:"name" json:"name"`       // "Bitcoin"

	// Market Data
	PriceUSD       float64 `db:"price_usd" json:"price_usd"`
	MarketCapUSD   float64 `db:"market_cap_usd" json:"market_cap_usd"`
	MarketCapRank  int     `db:"market_cap_rank" json:"market_cap_rank"`
	TotalVolumeUSD float64 `db:"total_volume_usd" json:"total_volume_usd"`

	// 24h Changes
	PriceChange24h            float64 `db:"price_change_24h" json:"price_change_24h"`
	PriceChangePercent24h     float64 `db:"price_change_percent_24h" json:"price_change_percent_24h"`
	MarketCapChange24h        float64 `db:"market_cap_change_24h" json:"market_cap_change_24h"`
	MarketCapChangePercent24h float64 `db:"market_cap_change_percent_24h" json:"market_cap_change_percent_24h"`

	// Supply
	CirculatingSupply float64  `db:"circulating_supply" json:"circulating_supply"`
	TotalSupply       *float64 `db:"total_supply" json:"total_supply,omitempty"`
	MaxSupply         *float64 `db:"max_supply" json:"max_supply,omitempty"`

	// All-Time High/Low
	ATH              float64   `db:"ath" json:"ath"`
	ATHChangePercent float64   `db:"ath_change_percent" json:"ath_change_percent"`
	ATHDate          time.Time `db:"ath_date" json:"ath_date"`
	ATL              float64   `db:"atl" json:"atl"`
	ATLChangePercent float64   `db:"atl_change_percent" json:"atl_change_percent"`
	ATLDate          time.Time `db:"atl_date" json:"atl_date"`

	// Metadata
	ImageURL    string    `db:"image_url" json:"image_url"`
	LastUpdated time.Time `db:"last_updated" json:"last_updated"`
	SnapshotAt  time.Time `db:"snapshot_at" json:"snapshot_at"` // When we captured this
}

// TableName for SQL
func (CoinSnapshot) TableName() string {
	return "coin_snapshots"
}

// JSONB support for complex fields if needed
type CoinMetadata struct {
	Categories  []string `json:"categories,omitempty"`
	Description string   `json:"description,omitempty"`
	Homepage    string   `json:"homepage,omitempty"`
}

// For storing JSONB in Postgres
func (m CoinMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *CoinMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}
