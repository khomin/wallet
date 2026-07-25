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

func DefaultTokenRegistry(tokens map[string][]config.TokenRegistry) *TokenRegistry {
	registry := NewTokenRegistry()

	// for chain, v := range tokens {
	// 	for _, token := range v {
	// 		registry.Register(entity.Token{
	// 			ID:       token.ID,
	// 			Chain:    strings.ToUpper(chain),
	// 			Symbol:   token.Symbol,
	// 			Name:     token.Name,
	// 			Address:  token.Address,
	// 			Decimals: token.Decimals,
	// 			IsNative: token.IsNative,
	// 		})
	// 	}
	// }
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

func (r *TokenRegistry) GetByAddress(chain, address string) (entity.Token, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	token, ok := r.tokensByAddress[fmt.Sprintf("%s:%s", chain, address)]
	return token, ok
}

func (r *TokenRegistry) GetByChainAndSymbol(chain, symbol string) (entity.Token, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	token, ok := r.tokensByChainSymbol[fmt.Sprintf("%s:%s", chain, symbol)]
	return token, ok
}

func (r *TokenRegistry) GetNative(chain string) (entity.Token, bool) {
	return r.GetByChainAndSymbol(chain, chain)
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

func (r *TokenRegistry) GetAllTokens() []entity.Token {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tokens := make([]entity.Token, 0, len(r.tokensByChainSymbol))
	for _, token := range r.tokensByChainSymbol {
		tokens = append(tokens, token)
	}
	return tokens
}
