package core

import (
	"fmt"
	"strings"
	"sync"
	"tracker/config"
	"tracker/internal/core/domain"
)

type TokenRegistry struct {
	mu            sync.RWMutex
	byName        map[string]config.TokenConfig
	byChainSymbol map[string]config.TokenConfig
	bySymbol      map[string]config.TokenConfig
}

func NewTokenRegistry() *TokenRegistry {
	return &TokenRegistry{
		byName:        make(map[string]config.TokenConfig),
		byChainSymbol: make(map[string]config.TokenConfig),
		bySymbol:      make(map[string]config.TokenConfig),
	}
}

func DefaultTokenRegistry(conf config.TokenRegistry) *TokenRegistry {
	registry := NewTokenRegistry()
	for _, token := range conf.Tokens {
		registry.Register(token)
	}
	return registry
}

func (r *TokenRegistry) Register(token config.TokenConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// index by name
	r.byName[token.Name] = token

	// index by symbol
	r.bySymbol[strings.ToUpper(token.Symbol)] = token

	// index by chain:symbol
	if token.Native != nil {
		symbolKey := fmt.Sprintf("%s:%s", strings.ToUpper(token.Native.Chain), strings.ToUpper(token.Symbol))
		r.byChainSymbol[symbolKey] = token
	} else {
		for chain := range token.Deployments {
			symbolKey := fmt.Sprintf("%s:%s", strings.ToUpper(chain), strings.ToUpper(token.Symbol))
			r.byChainSymbol[symbolKey] = token
		}
	}
}

func (r *TokenRegistry) GetByChainAndSymbol(chain, symbol string) (*domain.TokenChain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	token, ok := r.byChainSymbol[fmt.Sprintf("%s:%s", strings.ToUpper(chain), strings.ToUpper(symbol))]
	if ok {
		for depChain, dep := range token.Deployments {
			depChain = strings.ToUpper(depChain)
			if depChain == chain {
				return &domain.TokenChain{
					Symbol:  token.Symbol,
					Name:    token.Name,
					Chain:   depChain,
					Address: dep.Address,
				}, nil
			}
		}
		if token.Native != nil {
			if token.Native.Chain == chain {
				return &domain.TokenChain{
					Symbol:  token.Symbol,
					Name:    token.Name,
					Chain:   token.Native.Chain,
					Address: "",
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot find token")
}

func (r *TokenRegistry) GetBySymbol(symbol string) (domain.TokenRaw, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	symbol = strings.ToUpper(symbol)
	token, found := r.bySymbol[symbol]
	if !found {
		return domain.TokenRaw{}, false
	}
	chains := []string{}
	address := []string{}
	for chain, depl := range token.Deployments {
		chains = append(chains, chain)
		address = append(address, depl.Address)
	}
	return domain.TokenRaw{
		Symbol:   token.Symbol,
		Chains:   chains,
		Addrs:    address,
		IsNative: token.Native != nil,
		Name:     token.Name,
	}, true
}

func (r *TokenRegistry) GetByQuery(query string) []domain.TokenRaw {
	if query == "" {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToUpper(query)
	distinct := map[string]domain.TokenRaw{}
	for _, token := range r.byChainSymbol {
		matches := strings.Contains(strings.ToUpper(token.Name), query) || strings.Contains(token.Symbol, query)
		if matches {
			chains := make([]string, 0, len(token.Deployments))
			addrs := make([]string, 0, len(token.Deployments))
			for chain, depl := range token.Deployments {
				chains = append(chains, chain)
				addrs = append(addrs, depl.Address)
			}
			distinct[token.Name] = domain.TokenRaw{
				Symbol:   token.Symbol,
				Name:     token.Name,
				Chains:   chains,
				Addrs:    addrs,
				IsNative: token.Native != nil,
			}
		}
	}
	tokens := make([]domain.TokenRaw, 0, len(distinct))
	for _, token := range distinct {
		tokens = append(tokens, token)
	}
	return tokens
}
