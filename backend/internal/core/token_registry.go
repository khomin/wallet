package core

import (
	"fmt"
	"strings"
	"sync"
	"tracker/config"
	"tracker/internal/core/domain"
)

type Registry struct {
	mu                  sync.RWMutex
	tokensByName        map[string]config.TokenConfig
	tokensByChainSymbol map[string]config.TokenConfig
	tokensBySymbol      map[string]config.TokenConfig
}

func NewTokenRegistry() *Registry {
	return &Registry{
		tokensByName:        make(map[string]config.TokenConfig),
		tokensByChainSymbol: make(map[string]config.TokenConfig),
		tokensBySymbol:      make(map[string]config.TokenConfig),
	}
}

func DefaultTokenRegistry(conf config.TokenRegistry) *Registry {
	registry := NewTokenRegistry()
	for _, token := range conf.Tokens {
		registry.Register(token)
	}
	return registry
}

func (r *Registry) Register(token config.TokenConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// index by name
	r.tokensByName[token.Name] = token

	// index by symbol
	r.tokensBySymbol[strings.ToUpper(token.Symbol)] = token

	// index by chain:symbol
	if token.Native != nil {
		symbolKey := fmt.Sprintf("%s:%s", strings.ToUpper(token.Native.Chain), strings.ToUpper(token.Symbol))
		r.tokensByChainSymbol[symbolKey] = token
	} else {
		for chain := range token.Deployments {
			symbolKey := fmt.Sprintf("%s:%s", strings.ToUpper(chain), strings.ToUpper(token.Symbol))
			r.tokensByChainSymbol[symbolKey] = token
		}
	}
}

func (r *Registry) GetByChainAndSymbol(chain, symbol string) (*domain.TokenExactChain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	token, ok := r.tokensByChainSymbol[fmt.Sprintf("%s:%s", strings.ToUpper(chain), strings.ToUpper(symbol))]
	if ok {
		for depChain, dep := range token.Deployments {
			depChain = strings.ToUpper(depChain)
			if depChain == chain {
				return &domain.TokenExactChain{
					Symbol:   token.Symbol,
					Name:     token.Name,
					Chain:    depChain,
					Address:  dep.Address,
					IsNative: token.Native != nil,
				}, nil
			}
		}
		if token.Native != nil {
			if token.Native.Chain == chain {
				return &domain.TokenExactChain{
					Symbol:   token.Symbol,
					Name:     token.Name,
					Chain:    token.Native.Chain,
					Address:  "",
					IsNative: token.Native != nil,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot find token")
}

func (r *Registry) GetBySymbol(symbol string) (domain.Token, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	symbol = strings.ToUpper(symbol)
	token, found := r.tokensBySymbol[symbol]
	if !found {
		return domain.Token{}, false
	}
	chains := []string{}
	address := []string{}
	for chain, depl := range token.Deployments {
		chains = append(chains, chain)
		address = append(address, depl.Address)
	}
	return domain.Token{
		Symbol:   token.Symbol,
		Chains:   chains,
		Addrs:    address,
		IsNative: false,
		Name:     token.Name,
	}, true
}

func (r *Registry) GetByQuery(query string) []domain.Token {
	// r.mu.RLock()
	// defer r.mu.RUnlock()
	var tokens []domain.Token

	if query != "" {
		query = strings.ToUpper(query)
		// token, found := r.tokensByName[text]
		// if !found {
		// 	return nil
		// }
		for _, token := range r.tokensByChainSymbol {
			matches := strings.Contains(token.Name, query) || strings.Contains(token.Symbol, query)
			if matches {
				// token.Name
				// token.Native
				// token.Symbol
				// token.Deployments.key
				// token.Deployments[''].Address
				// token.Deployments[''].Decimals
				tokens = append(tokens, domain.Token{
					Symbol: token.Symbol,
					Name:   token.Name,
					Chains: []string{""},
					// Addrs:    token.,
					IsNative: false,
				})
			}
		}
	}
	return tokens
}

func (r *Registry) GetAllTokens() []domain.Token {
	tokens := []domain.Token{}
	// r.mu.RLock()
	// defer r.mu.RUnlock()

	// tokenUnique := map[string]domain.Token{}
	// for _, asset := range r.tokensByChainSymbol {
	// 	for _, token := range asset.Tokens {
	// 		tokenUnique[asset.Symbol] = token
	// 	}
	// }
	// for _, v := range tokenUnique {
	// 	tokens = append(tokens, v)
	// }
	return tokens
}
