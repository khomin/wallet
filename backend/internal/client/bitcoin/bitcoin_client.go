package bitcoin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BitcoinClient struct {
	baseURL    string
	httpClient *http.Client
}

type EsploraAddressStats struct {
	FundedTxoSum int64 `json:"funded_txo_sum"`
	SpentTxoSum  int64 `json:"spent_txo_sum"`
}

type EsploraAddressResponse struct {
	Address      string              `json:"address"`
	ChainStats   EsploraAddressStats `json:"chain_stats"`
	MempoolStats EsploraAddressStats `json:"mempool_stats"`
}

func NewBitcoinClient(host, user, password string) *BitcoinClient {
	return &BitcoinClient{
		baseURL: host,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *BitcoinClient) Connect(ctx context.Context) error {
	return nil
}

func (c *BitcoinClient) Close() {

}

func (c *BitcoinClient) GetBalance(ctx context.Context, address string) (float64, error) {
	url := fmt.Sprintf("%s/address/%s", c.baseURL, address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("esplora API returned status code: %d", resp.StatusCode)
	}

	var data EsploraAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// Calculate net balance: (Total Received - Total Spent) across chain + mempool
	funded := data.ChainStats.FundedTxoSum + data.MempoolStats.FundedTxoSum
	spent := data.ChainStats.SpentTxoSum + data.MempoolStats.SpentTxoSum
	satsBalance := funded - spent

	// Convert satoshis to BTC (1 BTC = 100,000,000 sats)
	return float64(satsBalance) / 1e8, nil
}

func (c *BitcoinClient) GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error) {
	return 0, nil
}
