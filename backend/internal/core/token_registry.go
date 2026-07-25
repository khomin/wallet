package core

import (
	"fmt"
	"strings"
	"sync"
	"tracker/config"
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

func DefaultTokenRegistry(config map[string][]config.TokenRegistry) *TokenRegistry {
	registry := NewTokenRegistry()

	assets, found := config["assets"]
	if found {
		for _, asset := range assets {
			for _, item := range asset.Items {
				symbol := asset.Symbol
				if item.Symbol != "" {
					symbol = item.Symbol
				}
				registry.Register(entity.Token{
					Name:     asset.Name,
					Symbol:   symbol,
					Chain:    strings.ToUpper(item.Chain),
					Address:  item.Address,
					Decimals: item.Decimals,
					IsNative: item.IsNative,
				})
			}
		}
	}
	return registry
}

func (r *TokenRegistry) Register(token entity.Token) {
	r.mu.Lock()
	defer r.mu.Unlock()

	chain := strings.ToUpper(token.Chain)
	symbol := strings.ToUpper(token.Symbol)
	address := strings.ToUpper(token.Address)

	// Index by chain:address
	addressKey := fmt.Sprintf("%s:%s", chain, address)
	r.tokensByAddress[addressKey] = token

	// Index by chain:symbol
	symbolKey := fmt.Sprintf("%s:%s", chain, symbol)
	r.tokensByChainSymbol[symbolKey] = token
}

func (r *TokenRegistry) GetByChainAndSymbol(chain, symbol string) (entity.Token, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	token, ok := r.tokensByChainSymbol[fmt.Sprintf("%s:%s", chain, symbol)]
	return token, ok
}

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

func (r *TokenRegistry) GetByText(text string) []entity.Token {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tokens []entity.Token
	for _, token := range r.tokensByChainSymbol {
		matches := strings.Contains(token.Name, text) || strings.Contains(token.Symbol, text)
		if matches {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func (r *TokenRegistry) GetAllTokens() []entity.Token {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tokenUnique := map[string]entity.Token{}
	for _, token := range r.tokensByChainSymbol {
		tokenUnique[token.Symbol] = token
	}
	tokens := []entity.Token{}
	for _, v := range tokenUnique {
		tokens = append(tokens, v)
	}
	return tokens
}
