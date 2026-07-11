package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CoinGeckoClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	coinLimit  int
}

func NewCoinGeckoClient(apiKey string) *CoinGeckoClient {
	return &CoinGeckoClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.coingecko.com/api/v3",
		apiKey:     apiKey,
		coinLimit:  250,
	}
}

func (c *CoinGeckoClient) GetMarket(ctx context.Context) ([]CoinGeckoCoin, error) {
	reqURL := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1&sparkline=false&locale=en", c.baseURL, c.coinLimit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Add("x-cg-pro-api-key", c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko returned status %s", resp.Status)
	}
	var coins []CoinGeckoCoin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, err
	}
	return coins, nil
}

func (c *CoinGeckoClient) GetCoinDetail(ctx context.Context, id string) (*CoinGeckoCoinDetail, error) {
	reqURL := fmt.Sprintf("%s/coins/%s", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Add("x-cg-pro-api-key", c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko returned status %s", resp.Status)
	}
	var coin CoinGeckoCoinDetail
	if err := json.NewDecoder(resp.Body).Decode(&coin); err != nil {
		return nil, err
	}
	return &coin, nil
}
