package domain

import (
	"math/big"
	"time"

	"github.com/google/uuid"
)

type Asset struct {
	Name     string  `json:"name"`
	Symbol   string  `json:"symbol"`
	Tokens   []Token `json:"tokens"`
	ImageURL string  `json:"image_url"`
}

type Token struct {
	Chain    string `json:"chain"`     // "ethereum"
	Symbol   string `json:"symbol"`    // "USDC"
	Name     string `json:"name"`      // "USD Coin"
	Address  string `json:"address"`   // "0xA0b8..."
	Decimals int    `json:"decimals"`  // 6
	IsNative bool   `json:"is_native"` // false (native is ETH)
	ImageURL string `json:"image_url"`
}

type TokenPrice struct {
	ID                             uuid.UUID `json:"id"`
	CoinID                         string    `json:"coin_id"`
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
	LastUpdated                    time.Time `json:"last_updated"`
}

// TokenBalance represents a balance of a specific token
type TokenBalance struct {
	Chain      string   `json:"chain"`
	Address    string   `json:"address"`     // User's wallet address
	Token      Token    `json:"token"`       // Token metadata
	Balance    *big.Int `json:"balance_wei"` // Raw balance
	BalanceDec float64  `json:"balance"`     // Human-readable
	PriceUSD   float64  `json:"price_usd"`
	ValueUSD   float64  `json:"value_usd"`
}

// AddressBalance represents all balances for a wallet address
type AddressBalance struct {
	Chain   string         `json:"chain"`
	Address string         `json:"address"`
	Native  *TokenBalance  `json:"native"` // ETH, BNB, etc.
	Tokens  []TokenBalance `json:"tokens"` // All ERC20 tokens
}

func (i *Token) AddImageURL(ImageURL string) {
	i.ImageURL = ImageURL
}
