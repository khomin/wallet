package domain

import (
	"math/big"
	"time"
)

type TokenMinimum struct {
	Symbol   string
	Name     string
	Chains   []string
	Addrs    []string
	IsNative bool
}

type TokenWithURL struct {
	Symbol   string
	Name     string
	Chains   []Chain
	Addrs    []string
	IsNative bool
	ImageURL string
}

// TODO: what does it do?
type TokenChain struct {
	Symbol  string
	Name    string
	Chain   string
	Address string
}

type TokenID struct {
	ID       string
	Symbol   string
	Name     string
	ImageURL string
}

type Chain struct {
	Symbol   string
	Name     string
	ImageUrl string
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
	Chain      string       `json:"chain"`
	Address    string       `json:"address"`     // User's wallet address
	Token      TokenMinimum `json:"token"`       // Token metadata
	Balance    *big.Int     `json:"balance_wei"` // Raw balance
	BalanceDec float64      `json:"balance"`     // Human-readable
	PriceUSD   float64      `json:"price_usd"`
	ValueUSD   float64      `json:"value_usd"`
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
