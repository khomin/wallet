package alchemy

import (
	"time"
)

type AlchemyPriceResponse struct {
	Data []TokenPriceData `json:"data"`
}

type TokenPriceData struct {
	Symbol string       `json:"symbol"`
	Prices []PriceEntry `json:"prices"`
	Error  *string      `json:"error,omitempty"`
}

type PriceEntry struct {
	Currency    string    `json:"currency"`
	Value       string    `json:"value"`
	LastUpdated time.Time `json:"last_updated"`
}
