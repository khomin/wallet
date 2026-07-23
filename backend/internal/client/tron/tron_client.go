package tron

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TronClient struct {
	grpcURL string
	apiKey  string
	client  *client.GrpcClient
}

func NewTronClient(grpcURL, apiKey string) *TronClient {
	return &TronClient{
		grpcURL: grpcURL,
		apiKey:  apiKey,
	}
}

func (c *TronClient) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil
	}
	if c.grpcURL == "" {
		return errors.New("tron grpc url is not configured")
	}
	grpcClient := client.NewGrpcClient(c.grpcURL)
	if c.apiKey != "" {
		_ = grpcClient.SetAPIKey(c.apiKey)
	}
	if err := grpcClient.Start(grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	c.client = grpcClient
	return nil
}

func (c *TronClient) Close() {
	if c.client != nil {
		c.client.Stop()
	}
}

func (c *TronClient) GetBalance(ctx context.Context, address string) (float64, error) {
	if err := c.Connect(ctx); err != nil {
		return 0, err
	}
	if c.client == nil {
		return 0, errors.New("tron client is not initialized")
	}
	account, err := c.client.GetAccountCtx(ctx, address)
	if err != nil {
		return 0, err
	}
	return float64(account.GetBalance()) / 1e6, nil
}

func (c *TronClient) GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error) {
	if err := c.Connect(ctx); err != nil {
		return 0, err
	}
	if c.client == nil {
		return 0, errors.New("tron client is not initialized")
	}
	// 1. Fetch raw token balance (*big.Int)
	rawBalance, err := c.client.TRC20ContractBalance(address, tokenAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch TRC20 balance: %w", err)
	}
	if rawBalance == nil {
		return 0, nil
	}

	// 2. Format balance with decimals (USDT uses 6 decimals)
	// Note: If you support arbitrary TRC20 tokens, you should query decimals dynamically or pass them in!
	decimals := 6.0

	balanceFloat := new(big.Float).SetInt(rawBalance)
	divisor := new(big.Float).SetFloat64(math.Pow(10, decimals))
	finalBalance, _ := new(big.Float).Quo(balanceFloat, divisor).Float64()

	return finalBalance, nil
}
