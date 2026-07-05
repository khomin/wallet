package models

import (
	"time"

	"github.com/google/uuid"
)

type CoinSnapshot struct {
	ID          uuid.UUID `db:"id" json:"id"`
	CoinID      string    `db:"coin_id" json:"coin_id"` // "bitcoin"
	Symbol      string    `db:"symbol" json:"symbol"`   // "btc"
	Name        string    `db:"name" json:"name"`       // "Bitcoin"
	ImageURL    string    `db:"image_url" json:"image_url"`
	LastUpdated time.Time `db:"last_updated" json:"last_updated"`
	SnapshotAt  time.Time `db:"snapshot_at" json:"snapshot_at"` // When we captured this
}

type PriceSnapshot struct {
	ID          uuid.UUID `db:"id" json:"id"`
	CoinID      string    `db:"coin_id" json:"coin_id"` // "bitcoin"
	Symbol      string    `db:"id" json:"symbol"`
	Name        string    `db:"name" json:"name"` // "Bitcoin"
	PriceUSD    float64   `db:"price_usd" json:"price_usd"`
	Change24h   float64   `db:"change_24h" json:"change_24h"`
	LastUpdated time.Time `db:"last_updated" json:"last_updated"`
}
