package solana

import (
	"context"
	"errors"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaClient struct {
	rpcURL string
	client *rpc.Client
}

func NewSolanaClient(rpcURL string) *SolanaClient {
	if rpcURL == "" {
		rpcURL = "https://api.mainnet-beta.solana.com"
	}
	return &SolanaClient{rpcURL: rpcURL}
}

func (c *SolanaClient) getClient() *rpc.Client {
	if c.client == nil {
		c.client = rpc.New(c.rpcURL)
	}
	return c.client
}

func (c *SolanaClient) GetBalance(ctx context.Context, address string) (float64, error) {
	if address == "" {
		return 0, errors.New("solana address is required")
	}
	client := c.getClient()
	publicKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return 0, err
	}
	result, err := client.GetBalance(ctx, publicKey, rpc.CommitmentFinalized)
	if err != nil {
		return 0, err
	}
	return float64(result.Value) / 1e9, nil
}
