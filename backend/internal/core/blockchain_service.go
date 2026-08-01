package core

import (
	"context"
	"fmt"
	"strings"

	"tracker/internal/client/bitcoin"
	"tracker/internal/client/ethereum"
	"tracker/internal/client/solana"
	"tracker/internal/client/tron"
)

type ChainProvider interface {
	GetBalance(ctx context.Context, address string) (float64, error)
	GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error)
	Connect(ctx context.Context) error
	Close()
}

type BlockchainService struct {
	providers     map[string]ChainProvider
	walletRepo    WalletRepository
	tokenRegistry *Registry
}

type AddressBalance struct {
	Chain   string
	Address string
	Balance float64
}

type BlockchainServiceDeps struct {
	EthMainnet    *ethereum.EthereumClient
	EthArbitrum   *ethereum.EthereumClient
	EthBase       *ethereum.EthereumClient
	Polygon       *ethereum.EthereumClient
	BNB           *ethereum.EthereumClient
	SOL           *solana.SolanaClient
	BTC           *bitcoin.BitcoinClient
	Tron          *tron.TronClient
	WalletRepo    WalletRepository
	TokenRegistry *Registry
}

func NewBlockchainService(deps BlockchainServiceDeps) *BlockchainService {
	providers := map[string]ChainProvider{}
	add := func(chain string, p ChainProvider) {
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
	return &BlockchainService{
		walletRepo:    deps.WalletRepo,
		tokenRegistry: deps.TokenRegistry,
		providers:     providers,
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
		balance, err := provider.GetBalance(ctx, address)
		if err != nil {
			return nil, err
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
		balance, err := provider.GetTokenBalance(ctx, address, token.Address)
		if err != nil {
			return nil, err
		}
		return &AddressBalance{
			Chain:   chain,
			Address: token.Address,
			Balance: balance,
		}, nil
	}
}
