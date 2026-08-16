package domain

import (
	"math/big"
	"time"
)

type TokenRaw struct {
	Symbol   string   `json:"symbol"`  // e.g., "USDT"
	Name     string   `json:"name"`    // e.g., "Tether USD"
	Chains   []string `json:"chain"`   // e.g., "TRX"
	Addrs    []string `json:"address"` // "native" or "0x..."
	IsNative bool     `json:"is_native"`
}

type TokenWithURL struct {
	TokenRaw
	ImageURL string `json:"image_url"`
}

type TokenChain struct {
	Symbol  string `json:"symbol"`  // e.g., "USDT"
	Name    string `json:"name"`    // e.g., "Tether USD"
	Chain   string `json:"chain"`   // e.g., "TRX"
	Address string `json:"address"` // "native" or "0x..."
}

type TokenID struct {
	ID       string `json:"id"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

type TokenPrice struct {
	ID                             string    `json:"id"`
	Name                           string    `json:"name"`
	Symbol                         string    `json:"symbol"`
	CurrentPrice                   float64   `json:"current_price"`
	Change_24h                     float64   `json:"change_24h"`
	MarketCap                      float64   `json:"market_cap"`
	TotalVolume                    float64   `json:"total_volume"`
	High_24h                       float64   `json:"high_24h"`
	Low_24h                        float64   `json:"low_24h"`
	PriceChange_24h                float64   `json:"price_change_24h"`
	PriceChangePercentage_24h      float64   `json:"price_change_percentage_24h"`
	MarketCapChange_24h            float64   `json:"market_cap_change_24h"`
	MarketCapChange_percentage_24h float64   `json:"market_cap_change_percentage_24h"`
	UpdatedAt                      time.Time `json:"updated_at"`
}

type TokenBalance struct {
	Chain      string   `json:"chain"`
	Address    string   `json:"address"`     // User's wallet address
	Token      TokenRaw `json:"token"`       // Token metadata
	Balance    *big.Int `json:"balance_wei"` // Raw balance
	BalanceDec float64  `json:"balance"`     // Human-readable
	PriceUSD   float64  `json:"price_usd"`
	ValueUSD   float64  `json:"value_usd"`
}

type AddressBalance struct {
	Chain   string         `json:"chain"`
	Address string         `json:"address"`
	Native  *TokenBalance  `json:"native"` // ETH, BNB, etc.
	Tokens  []TokenBalance `json:"tokens"` // All ERC20 tokens
}

func (t *TokenPrice) LessThanOrEqual(price float64) bool {
	return t.CurrentPrice <= price
}

func (t *TokenPrice) GreaterThanOrEqual(price float64) bool {
	return t.CurrentPrice >= price
}
