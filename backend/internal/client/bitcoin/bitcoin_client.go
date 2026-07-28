package bitcoin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
			DisableTLS:   true,
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
	if c.client == nil {
		return 0, errors.New("bitcoin client is not initialized")
	}
	if err := c.Connect(ctx); err != nil {
		return 0, err
	}
	err := c.EnsureWalletLoaded(ctx)
	if err != nil {
		return 0, err
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

func (c *BitcoinClient) EnsureWalletLoaded(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("bitcoin client is nil")
	}
	// 1. Call `listwallets` to see if a wallet is already active
	resp, err := c.client.RawRequest("listwallets", nil)
	if err == nil {
		var wallets []string
		if err := json.Unmarshal(resp, &wallets); err == nil && len(wallets) > 0 {
			return nil // Wallet is already loaded!
		}
	}
	// 2. Try `loadwallet` in case "default" exists on disk
	walletNameParam, _ := json.Marshal("default")
	_, err = c.client.RawRequest("loadwallet", []json.RawMessage{walletNameParam})
	if err == nil {
		return nil // Loaded successfully!
	}
	// 3. If loading failed (doesn't exist yet), create it via `createwallet`
	// Parameters: [wallet_name, disable_private_keys, blank, passphrase, avoid_reuse, descriptors, load_on_startup]
	_, err = c.client.RawRequest("createwallet", []json.RawMessage{walletNameParam})
	if err != nil {
		return fmt.Errorf("failed to create wallet 'default': %w", err)
	}
	return nil
}
