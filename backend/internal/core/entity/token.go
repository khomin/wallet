package entity

import "math/big"

type Token struct {
	ID       string `json:"id"`        // "usdc_ethereum"
	Chain    string `json:"chain"`     // "ethereum"
	Symbol   string `json:"symbol"`    // "USDC"
	Name     string `json:"name"`      // "USD Coin"
	Address  string `json:"address"`   // "0xA0b8..."
	Decimals int    `json:"decimals"`  // 6
	IsNative bool   `json:"is_native"` // false (native is ETH)
	LogoURL  string `json:"logo_url"`
}

type ChainTokens struct {
	Chain  string  `json:"chain"`
	Native Token   `json:"native"` // ETH, BNB, MATIC, etc.
	Tokens []Token `json:"tokens"` // All ERC20 tokens
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
