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

type BlockchainService struct {
	Ethereum *ethereum.EthereumClient
	Solana   *solana.SolanaClient
	Bitcoin  *bitcoin.BitcoinClient
	Tron     *tron.TronClient
}

type Balance struct {
	Chain   string
	Address string
	Balance float64
}

func NewBlockchainService(
	eth *ethereum.EthereumClient,
	sol *solana.SolanaClient,
	btc *bitcoin.BitcoinClient,
	tronCli *tron.TronClient,
) *BlockchainService {
	return &BlockchainService{
		Ethereum: eth,
		Solana:   sol,
		Bitcoin:  btc,
		Tron:     tronCli,
	}
}

func (s *BlockchainService) ConnectAll(ctx context.Context) error {
	if s.Ethereum != nil {
		if err := s.Ethereum.Connect(ctx); err != nil {
			return fmt.Errorf("ethereum connect: %w", err)
		}
	}
	if s.Solana != nil {
		// No explicit connection step required, just validate endpoint.
	}
	if s.Bitcoin != nil {
		if err := s.Bitcoin.Connect(ctx); err != nil {
			return fmt.Errorf("bitcoin connect: %w", err)
		}
	}
	if s.Tron != nil {
		if err := s.Tron.Connect(ctx); err != nil {
			return fmt.Errorf("tron connect: %w", err)
		}
	}
	return nil
}

func (s *BlockchainService) GetBalance(
	ctx context.Context,
	chain string,
	address string,
) (*Balance, error) {
	balance := Balance{
		Chain:   chain,
		Address: address,
		Balance: 0,
	}
	switch strings.ToLower(chain) {
	case "ethereum", "eth", "bnb", "polygon", "arbitrum", "base":
		if s.Ethereum == nil {
			return nil, fmt.Errorf("ethereum provider not configured")
		}
		v, err := s.Ethereum.GetBalance(ctx, address)
		if err != nil {
			return nil, err
		}
		balance.Balance = v
	case "solana", "sol":
		if s.Solana == nil {
			return nil, fmt.Errorf("solana provider not configured")
		}
		v, err := s.Solana.GetBalance(ctx, address)
		if err != nil {
			return nil, err
		}
		balance.Balance = v
	case "bitcoin", "btc":
		if s.Bitcoin == nil {
			return nil, fmt.Errorf("bitcoin provider not configured")
		}
		v, err := s.Bitcoin.GetBalance(ctx, address)
		if err != nil {
			return nil, err
		}
		balance.Balance = v
	case "tron":
		if s.Tron == nil {
			return nil, fmt.Errorf("tron provider not configured")
		}
		v, err := s.Tron.GetBalance(ctx, address)
		if err != nil {
			return nil, err
		}
		balance.Balance = v
	default:
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
	return &balance, nil
}
