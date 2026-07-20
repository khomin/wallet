package bitcoin

import (
	"context"
	"errors"

	btcaddress "github.com/btcsuite/btcd/address/v2"
	"github.com/btcsuite/btcd/chaincfg/v2"
	"github.com/btcsuite/btcd/rpcclient"
)

type BitcoinClient struct {
	cfg    rpcclient.ConnConfig
	client *rpcclient.Client
}

func NewBitcoinClient(host, user, password string) *BitcoinClient {
	return &BitcoinClient{
		cfg: rpcclient.ConnConfig{
			Host:         host,
			User:         user,
			Pass:         password,
			HTTPPostMode: true,
		},
	}
}

func (c *BitcoinClient) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil
	}
	if c.cfg.Host == "" {
		return errors.New("bitcoin rpc host is not configured")
	}
	client, err := rpcclient.New(&c.cfg, nil)
	if err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *BitcoinClient) Close() {
	if c.client != nil {
		c.client.Shutdown()
	}
}

func (c *BitcoinClient) GetBalance(ctx context.Context, address string) (float64, error) {
	if err := c.Connect(ctx); err != nil {
		return 0, err
	}
	if c.client == nil {
		return 0, errors.New("bitcoin client is not initialized")
	}
	addr, err := btcaddress.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		return 0, err
	}
	amount, err := c.client.GetReceivedByAddress(addr)
	if err != nil {
		return 0, err
	}
	return amount.ToBTC(), nil
}

func (c *BitcoinClient) GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error) {
	return 0, nil
}
