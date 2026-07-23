package tron

import (
	"context"
	"errors"

	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TronClient struct {
	grpcURL string
	apiKey  string
	client  *client.GrpcClient
}

// TODO: TRON: github.com/fbsobreira/gotron-sdk
func NewTronClient(grpcURL, apiKey string) *TronClient {
	if grpcURL == "" {
		grpcURL = "grpc.trongrid.io:50051"
	}
	return &TronClient{grpcURL: grpcURL, apiKey: apiKey}
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
	// TODO: tokens
	return 0, nil
}
