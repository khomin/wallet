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
	tokensByName        map[string]entity.Asset
	tokensByChainSymbol map[string]entity.Asset
}

func NewTokenRegistry() *TokenRegistry {
	return &TokenRegistry{
		tokensByName:        make(map[string]entity.Asset),
		tokensByChainSymbol: make(map[string]entity.Asset),
	}
}

func DefaultTokenRegistry(config map[string][]config.TokenRegistry) *TokenRegistry {
	registry := NewTokenRegistry()

	assets, found := config["assets"]
	if found {
		for _, asset := range assets {
			out1 := entity.Asset{}
			for _, item := range asset.Items {
				out1.Symbol = asset.Symbol
				out1.Name = asset.Name
				out1.Tokens = append(out1.Tokens, entity.Token{
					Name:     asset.Name,
					Symbol:   item.Symbol,
					Chain:    strings.ToUpper(item.Chain),
					Address:  item.Address,
					Decimals: item.Decimals,
					IsNative: item.IsNative,
				})
			}
			registry.Register(out1)
		}
	}
	return registry
}

func (r *TokenRegistry) Register(asset entity.Asset) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// - id: "tether-gold"
	//   name: "Tether Gold"
	//   symbol: "XAUt"
	//   items:
	//     - chain: "ETH"
	//       symbol: "XAUt"
	//       address: "0x68749665FF8D2d112Fa859AA293F07A622782F38"
	//       decimals: 6
	//       is_native: false
	//     - chain: "SOL"
	//       symbol: "XAUt0"
	//       address: "AymATz4TCL9sWNEEV9Kvyz45CHVhDZ6kUgjTJPzLpU9P"
	//       decimals: 18
	//       is_native: false

	// Index by name
	r.tokensByName[asset.Name] = asset

	for idx, token := range asset.Tokens {
		chain := strings.ToUpper(token.Chain)
		if token.Symbol == "" {
			if asset.Symbol != "" {
				token.Symbol = strings.ToUpper(asset.Symbol)
			} else {
				token.Symbol = chain
			}
			asset.Tokens[idx] = token
		}
		symbol := strings.ToUpper(token.Symbol)
		// Index by chain:symbol
		symbolKey := fmt.Sprintf("%s:%s", chain, symbol)
		r.tokensByChainSymbol[symbolKey] = asset
	}
}

func (r *TokenRegistry) GetByChainAndSymbol(chain, symbol string) (string, entity.Token, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	token, ok := r.tokensByChainSymbol[fmt.Sprintf("%s:%s", strings.ToUpper(chain), strings.ToUpper(symbol))]
	if ok && len(token.Tokens) != 0 {
		for _, i := range token.Tokens {
			if i.Chain == chain {
				return token.Symbol, i, true
			}
		}
		return token.Symbol, token.Tokens[0], ok
	}
	return "", entity.Token{}, false
}

func (r *TokenRegistry) GetBySymbol(symbol string) (entity.Token, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	symbol = strings.ToUpper(symbol)
	for _, i := range r.tokensByChainSymbol {
		for _, token := range i.Tokens {
			if strings.ToUpper(i.Symbol) == symbol {
				return token, true
			}
		}
	}
	return entity.Token{}, false
}

func (r *TokenRegistry) GetAssetsByText(text string) []entity.Asset {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tokens []entity.Asset
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
	for _, asset := range r.tokensByChainSymbol {
		for _, token := range asset.Tokens {
			tokenUnique[asset.Symbol] = token
		}
	}
	tokens := []entity.Token{}
	for _, v := range tokenUnique {
		tokens = append(tokens, v)
	}
	return tokens
}
