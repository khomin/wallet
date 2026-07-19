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
	tokenRegistry *TokenRegistry
}

type AddressBalance struct {
	Chain   string
	Address string
	Balance float64
}

func NewBlockchainService(
	ethereumMainnet *ethereum.EthereumClient,
	ethereumArbitrum *ethereum.EthereumClient,
	ethereumBase *ethereum.EthereumClient,
	polygon *ethereum.EthereumClient,
	bnb *ethereum.EthereumClient,
	sol *solana.SolanaClient,
	btc *bitcoin.BitcoinClient,
	tron *tron.TronClient,
	walletRepo WalletRepository,
	tokenRegistry *TokenRegistry,
) *BlockchainService {
	return &BlockchainService{
		providers: map[string]ChainProvider{
			"ETH":     ethereumMainnet,
			"ARB":     ethereumArbitrum,
			"BASE":    ethereumBase,
			"POLYGON": polygon,
			"BNB":     bnb,
			"BSC":     bnb,
			"SOL":     sol,
			"TRX":     tron,
		},
		walletRepo:    walletRepo,
		tokenRegistry: tokenRegistry,
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
		token, ok := s.tokenRegistry.GetByChainAndSymbol(chain, tokenSymbol)
		if !ok {
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
