package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	coins := []CoinGeckoCoin{}
	// basic coins
	params := url.Values{}
	params.Add("vs_currency", "usd")
	params.Add("order", "market_cap_desc")
	params.Add("per_page", strconv.Itoa(c.coinLimit))
	if v, err := c.requestMarket(ctx, params, false); err == nil {
		coins = append(coins, v...)
	} else {
		return nil, err
	}
	// stocks
	params = url.Values{}
	params.Add("vs_currency", "usd")
	params.Add("order", "market_cap_desc")
	params.Add("category", "tokenized-stock")
	params.Add("per_page", strconv.Itoa(c.coinLimit))
	if v, err := c.requestMarket(ctx, params, false); err == nil {
		coins = append(coins, v...)
	} else {
		return nil, err
	}
	return coins, nil
}

func (c *CoinGeckoClient) requestMarket(ctx context.Context, params url.Values, useApiKey bool) ([]CoinGeckoCoin, error) {
	reqBase, err := url.Parse(fmt.Sprintf("%s/coins/markets", c.baseURL))
	if err != nil {
		return nil, err
	}
	reqBase.RawQuery = params.Encode()

	urlTest := reqBase.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlTest, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" && useApiKey {
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
