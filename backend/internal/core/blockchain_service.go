package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"tracker/internal/client"

	"github.com/ethereum/go-ethereum/rpc"
)

var (
	ErrProviderTimeout     = errors.New("provider timeout")
	ErrProviderRateLimit   = errors.New("provider rate limit")
	ErrProviderUnavailable = errors.New("provider unavailable")
)

type BlockchainService struct {
	providers     map[string]*client.RateLimitedProvider
	walletRepo    WalletRepository
	tokenRegistry *TokenRegistry
	cache         PriceCache
}

type AddressBalance struct {
	Chain   string
	Address string
	Balance float64
}

type BlockchainServiceDeps struct {
	EthMainnet    *client.RateLimitedProvider
	EthArbitrum   *client.RateLimitedProvider
	EthBase       *client.RateLimitedProvider
	Polygon       *client.RateLimitedProvider
	BNB           *client.RateLimitedProvider
	SOL           *client.RateLimitedProvider
	BTC           *client.RateLimitedProvider
	Tron          *client.RateLimitedProvider
	Ripple        *client.RateLimitedProvider
	WalletRepo    WalletRepository
	TokenRegistry *TokenRegistry
	Cache         PriceCache
}

func NewBlockchainService(deps BlockchainServiceDeps) *BlockchainService {
	providers := map[string]*client.RateLimitedProvider{}
	add := func(chain string, p *client.RateLimitedProvider) {
		if p != nil {
			providers[chain] = p
		}
	}
	add("BTC", deps.BTC)
	add("ETH", deps.EthMainnet)
	add("ARB", deps.EthArbitrum)
	add("BASE", deps.EthBase)
	add("POL", deps.Polygon)
	add("BNB", deps.BNB)
	add("BSC", deps.BNB)
	add("SOL", deps.SOL)
	add("TRX", deps.Tron)
	add("XRP", deps.Ripple)
	return &BlockchainService{
		walletRepo:    deps.WalletRepo,
		tokenRegistry: deps.TokenRegistry,
		providers:     providers,
		cache:         deps.Cache,
	}
}

func (s *BlockchainService) ConnectAll(ctx context.Context) error {
	for key, value := range s.providers {
		if err := value.Connect(ctx); err != nil {
			return fmt.Errorf("%s connect: %w", key, err)
		}
	}
	return nil
}

func (s *BlockchainService) GetBalance(ctx context.Context, chain string, address string, tokenSymbol string) (*AddressBalance, error) {
	chain = strings.ToUpper(chain)
	provider, found := s.providers[chain]
	if !found {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
	if chain == tokenSymbol {
		if val, err := s.cache.GetBalanceNative(ctx, chain, address); err == nil {
			return &AddressBalance{
				Chain:   chain,
				Address: address,
				Balance: val,
			}, nil
		}
		balance, err := provider.GetBalance(ctx, address)
		if err != nil {
			var rpcErr rpc.HTTPError
			if errors.As(err, &rpcErr) {
				if rpcErr.StatusCode == 408 || rpcErr.StatusCode == 429 {
					return nil, ErrProviderRateLimit
				}
			}
			return nil, err
		}
		if err = s.cache.SetBalanceNative(ctx, chain, address, balance); err != nil {
			slog.Error("failed to cache balance", "error", err)
		}
		return &AddressBalance{
			Chain:   chain,
			Address: address,
			Balance: balance,
		}, nil
	} else {
		token, err := s.tokenRegistry.GetByChainAndSymbol(chain, tokenSymbol)
		if err != nil {
			return nil, fmt.Errorf("token not found %s", tokenSymbol)
		}
		if val, err := s.cache.GetBalanceToken(ctx, chain, address, tokenSymbol); err == nil {
			return &AddressBalance{
				Chain:   chain,
				Address: token.Address,
				Balance: val,
			}, nil
		}
		balance, err := provider.GetTokenBalance(ctx, address, token.Address)
		if err != nil {
			return nil, err
		}
		if err = s.cache.SetBalanceToken(ctx, chain, address, tokenSymbol, balance); err != nil {
			slog.Error("failed to cache balance", "error", err)
		}
		return &AddressBalance{
			Chain:   chain,
			Address: token.Address,
			Balance: balance,
		}, nil
	}
}

func (s *BlockchainService) ValidateAddress(chain string, address string, tokenSymbol string) error {
	chain = strings.ToUpper(chain)
	provider, found := s.providers[chain]
	if !found {
		return fmt.Errorf("unsupported chain: %s", chain)
	}
	if chain == tokenSymbol {
		err := provider.ValidateAddress(address, address)
		if err != nil {
			return err
		}
		return nil
	} else {
		token, err := s.tokenRegistry.GetByChainAndSymbol(chain, tokenSymbol)
		if err != nil {
			return fmt.Errorf("token not found %s", tokenSymbol)
		}
		err = provider.ValidateAddress(address, token.Address)
		if err != nil {
			return err
		}
		return nil
	}
}
