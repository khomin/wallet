package ethereum

import (
	"context"
	"errors"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthereumClient struct {
	rpcURL  string
	client  *ethclient.Client
	timeout time.Duration
}

func NewEthereumClient(rpcURL string) *EthereumClient {
	return &EthereumClient{
		rpcURL:  rpcURL,
		timeout: 10 * time.Second,
	}
}

func (c *EthereumClient) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil
	}
	if c.rpcURL == "" {
		return errors.New("ethereum rpc url is not configured")
	}
	client, err := ethclient.DialContext(ctx, c.rpcURL)
	if err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *EthereumClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

func (c *EthereumClient) GetBalance(ctx context.Context, address string) (float64, error) {
	if err := c.Connect(ctx); err != nil {
		return 0, err
	}
	if c.client == nil {
		return 0, errors.New("ethereum client is not initialized")
	}
	account := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, account, nil)
	if err != nil {
		return 0, err
	}
	balanceFloat := new(big.Float).SetInt(balance)
	etherValue := new(big.Float).Quo(balanceFloat, big.NewFloat(math.Pow10(18)))
	f, _ := etherValue.Float64()
	return f, nil
}
