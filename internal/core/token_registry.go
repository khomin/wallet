package core

import (
	"fmt"
	"sync"
	"tracker/internal/core/entity"
)

type TokenRegistry struct {
	mu                  sync.RWMutex
	tokensByAddress     map[string]entity.Token // key: "chain:address"
	tokensByChainSymbol map[string]entity.Token // key: "chain:symbol"
}

func NewTokenRegistry() *TokenRegistry {
	return &TokenRegistry{
		tokensByAddress:     make(map[string]entity.Token),
		tokensByChainSymbol: make(map[string]entity.Token),
	}
}

// Register adds a token to the registry
func (r *TokenRegistry) Register(token entity.Token) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Index by chain:address
	addressKey := fmt.Sprintf("%s:%s", token.Chain, token.Address)
	r.tokensByAddress[addressKey] = token

	// Index by chain:symbol
	symbolKey := fmt.Sprintf("%s:%s", token.Chain, token.Symbol)
	r.tokensByChainSymbol[symbolKey] = token
}

// GetByAddress returns a token by chain and contract address
func (r *TokenRegistry) GetByAddress(chain, address string) (entity.Token, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	token, ok := r.tokensByAddress[fmt.Sprintf("%s:%s", chain, address)]
	return token, ok
}

// GetByChainAndSymbol returns a token by chain and symbol
func (r *TokenRegistry) GetByChainAndSymbol(chain, symbol string) (entity.Token, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	token, ok := r.tokensByChainSymbol[fmt.Sprintf("%s:%s", chain, symbol)]
	return token, ok
}

// GetNative returns the native token for a chain
func (r *TokenRegistry) GetNative(chain string) (entity.Token, bool) {
	return r.GetByChainAndSymbol(chain, "native")
}

// GetAllByChain returns all tokens for a specific chain
func (r *TokenRegistry) GetAllByChain(chain string) []entity.Token {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tokens []entity.Token
	for _, token := range r.tokensByChainSymbol {
		if token.Chain == chain {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func DefaultTokenRegistry() *TokenRegistry {
	registry := NewTokenRegistry()

	// Ethereum Native
	registry.Register(entity.Token{
		ID:       "eth_ethereum",
		Chain:    "ETH",
		Symbol:   "ETH",
		Name:     "Ethereum",
		Address:  "native",
		Decimals: 18,
		IsNative: true,
		// CoingeckoID: "ethereum",
	})

	// Ethereum Tokens
	registry.Register(entity.Token{
		ID:       "usdc_ethereum",
		Chain:    "ETH",
		Symbol:   "USDC",
		Name:     "USD Coin",
		Address:  "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		Decimals: 6,
		IsNative: false,
		// CoingeckoID: "usd-coin",
	})

	registry.Register(entity.Token{
		ID:       "usdt_ethereum",
		Chain:    "ETH",
		Symbol:   "USDT",
		Name:     "Tether USD",
		Address:  "0xdAC17F958D2ee523a2206206994597C13D831ec7",
		Decimals: 6,
		IsNative: false,
		// CoingeckoID: "tether",
	})

	registry.Register(entity.Token{
		ID:       "wbtc_ethereum",
		Chain:    "ETH",
		Symbol:   "WBTC",
		Name:     "Wrapped Bitcoin",
		Address:  "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
		Decimals: 8,
		IsNative: false,
		// CoingeckoID: "wrapped-bitcoin",
	})

	registry.Register(entity.Token{
		ID:       "xaut_ethereum",
		Chain:    "ETH",
		Symbol:   "XAUT",
		Name:     "Tether Gold",
		Address:  "0x68749665FF8D2d112Fa859AA293F07A622782F38",
		Decimals: 6,
		IsNative: false,
		// CoingeckoID: "tether-gold",
	})

	// ... Add more tokens ...

	// Polygon Native
	registry.Register(entity.Token{
		ID:       "pol_polygon",
		Chain:    "POL",
		Symbol:   "POL",
		Name:     "Polygon",
		Address:  "native",
		Decimals: 18,
		IsNative: true,
		// CoingeckoID: "polygon",
	})

	// Polygon Tokens
	registry.Register(entity.Token{
		ID:       "usdc_polygon",
		Chain:    "POL",
		Symbol:   "USDC",
		Name:     "USD Coin (Polygon)",
		Address:  "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359",
		Decimals: 6,
		IsNative: false,
		// CoingeckoID: "usd-coin",
	})
	// PAXG (Paxos Gold) - Ethereum
	registry.Register(entity.Token{
		ID:       "paxg_ethereum",
		Chain:    "ETH",
		Symbol:   "PAXG",
		Name:     "Paxos Gold",
		Address:  "0x45804880De22913dAFE09f4980848ECE6EcbAf78",
		Decimals: 18,
		IsNative: false,
	})
	return registry
}

func (r *TokenRegistry) GetAllTokens() []entity.Token {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tokens := make([]entity.Token, 0, len(r.tokensByChainSymbol))
	for _, token := range r.tokensByChainSymbol {
		tokens = append(tokens, token)
	}
	return tokens
}
