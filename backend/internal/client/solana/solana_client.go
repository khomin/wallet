package solana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaClient struct {
	rpcURL string
	client *rpc.Client
}

func NewSolanaClient(rpcURL string) *SolanaClient {
	return &SolanaClient{rpcURL: rpcURL}
}

func (c *SolanaClient) getClient() *rpc.Client {
	if c.client == nil {
		c.client = rpc.New(c.rpcURL)
	}
	return c.client
}

func (c *SolanaClient) Connect(ctx context.Context) error {
	// solana-go's RPC client is HTTP-based and doesn't require a persistent dial handshake
	// unless using WebSockets.
	return nil
}

func (c *SolanaClient) Close() {
	// HTTP client doesn't need explicit closure, but we leave the signature
	// to satisfy your shared interfaces if needed.
}

func (c *SolanaClient) GetBalance(ctx context.Context, address string) (float64, error) {
	if address == "" {
		return 0, errors.New("solana address is required")
	}

	client := c.getClient()
	publicKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return 0, fmt.Errorf("invalid solana address: %w", err)
	}

	result, err := client.GetBalance(ctx, publicKey, rpc.CommitmentFinalized)
	if err != nil {
		return 0, err
	}

	// 1 SOL = 1 Billion Lamports (1e9)
	return float64(result.Value) / 1e9, nil
}

// GetTokenBalance queries all token accounts owned by a wallet for a specific SPL token.
func (c *SolanaClient) GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error) {
	if address == "" || tokenAddress == "" {
		return 0, errors.New("address and tokenAddress are required")
	}

	client := c.getClient()

	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return 0, fmt.Errorf("invalid wallet address: %w", err)
	}

	tokenMint, err := solana.PublicKeyFromBase58(tokenAddress)
	if err != nil {
		return 0, fmt.Errorf("invalid token address: %w", err)
	}

	// 1. Ask the node for all accounts owned by the user that hold this specific token
	conf := &rpc.GetTokenAccountsConfig{
		Mint: &tokenMint,
	}

	// 2. Request JSONParsed encoding. This is the secret sauce:
	// The RPC will automatically calculate decimals and give us the human-readable UI amount.
	opts := &rpc.GetTokenAccountsOpts{
		Encoding: solana.EncodingJSONParsed,
	}

	out, err := client.GetTokenAccountsByOwner(ctx, pubKey, conf, opts)
	if err != nil {
		return 0, err
	}

	// Create an inline struct to easily unmarshal the nested RPC JSON response
	type parsedData struct {
		Parsed struct {
			Info struct {
				TokenAmount struct {
					UIAmount float64 `json:"uiAmount"`
				} `json:"tokenAmount"`
			} `json:"info"`
		} `json:"parsed"`
	}

	var totalBalance float64

	// 3. Loop through all accounts and sum their parsed balances
	for _, acc := range out.Value {
		rawJSON := acc.Account.Data.GetRawJSON()
		if len(rawJSON) > 0 {
			var p parsedData
			if err := json.Unmarshal(rawJSON, &p); err == nil {
				totalBalance += p.Parsed.Info.TokenAmount.UIAmount
			}
		}
	}

	return totalBalance, nil
}
