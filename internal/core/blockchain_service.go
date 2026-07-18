package core

import (
	"context"
	"fmt"

	"tracker/internal/client/bitcoin"
	"tracker/internal/client/ethereum"
	"tracker/internal/client/solana"
	"tracker/internal/client/tron"
)

type ChainProvider interface {
	GetBalance(ctx context.Context, address string) (float64, error)
	Connect(ctx context.Context) error
	Close()
}

type BlockchainService struct {
	providers  map[string]ChainProvider
	walletRepo WalletRepository
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
) *BlockchainService {
	return &BlockchainService{
		providers: map[string]ChainProvider{
			"ethereum": ethereumMainnet,
			"eth":      ethereumMainnet,
			"arbitrum": ethereumArbitrum,
			"arb":      ethereumArbitrum,
			"base":     ethereumBase,
			"polygon":  polygon,
			"bnb":      bnb,
			"bsc":      bnb,
			"solana":   sol,
			"sol":      sol,
			"tron":     tron,
			"trx":      tron,
		},
		walletRepo: walletRepo,
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

func (s *BlockchainService) GetBalance(ctx context.Context, chain string, address string) (*AddressBalance, error) {
	provider, found := s.providers[chain]
	if !found {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
	balance, err := provider.GetBalance(ctx, address)
	if err != nil {
		return nil, err
	}
	return &AddressBalance{
		Chain:   chain,
		Address: address,
		Balance: balance,
	}, nil
}
