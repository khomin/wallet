package ripple

import (
	"context"
	"time"
)

// TODO: sdk ripple rubblelabs/ripple blinklabs-io/gouroboros

type RippleClient struct {
	rpcURL string
	// client  *ethclient.Client
	timeout time.Duration
}

func NewRippleClient(rpcURL string) *RippleClient {
	return &RippleClient{
		rpcURL:  rpcURL,
		timeout: 10 * time.Second,
	}
}

func (c *RippleClient) Connect(ctx context.Context) error {
	// if c.client != nil {
	// 	return nil
	// }
	// if c.rpcURL == "" {
	// 	return errors.New("ethereum rpc url is not configured")
	// }
	// client, err := ethclient.DialContext(ctx, c.rpcURL)
	// if err != nil {
	// 	return err
	// }
	// c.client = client
	return nil
}

func (c *RippleClient) Close() {
	// if c.client != nil {
	// 	c.client.Close()
	// }
}

func (c *RippleClient) GetBalance(ctx context.Context, address string) (float64, error) {
	// if err := c.Connect(ctx); err != nil {
	// 	return 0, err
	// }
	// if c.client == nil {
	// 	return 0, errors.New("ripple client is not initialized")
	// }
	// account := common.HexToAddress(address)
	// balance, err := c.client.BalanceAt(ctx, account, nil)
	// if err != nil {
	// 	return 0, err
	// }
	// balanceFloat := new(big.Float).SetInt(balance)
	// etherValue := new(big.Float).Quo(balanceFloat, big.NewFloat(math.Pow10(18)))
	// f, _ := etherValue.Float64()
	// return f, nil
	return 0, nil
}

func (c *RippleClient) GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error) {
	return 0, nil
}
