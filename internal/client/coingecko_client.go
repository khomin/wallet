package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CoinGeckoCoin - maps exactly to the CoinGecko API response
type CoinGeckoCoin struct {
	ID                        string      `json:"id"`
	Symbol                    string      `json:"symbol"`
	Name                      string      `json:"name"`
	Image                     string      `json:"image"`
	CurrentPrice              float64     `json:"current_price"`
	MarketCap                 float64     `json:"market_cap"`
	MarketCapRank             int         `json:"market_cap_rank"`
	FullyDilutedValuation     float64     `json:"fully_diluted_valuation"`
	TotalVolume               float64     `json:"total_volume"`
	High24h                   float64     `json:"high_24h"`
	Low24h                    float64     `json:"low_24h"`
	PriceChange24h            float64     `json:"price_change_24h"`
	PriceChangePercent24h     float64     `json:"price_change_percentage_24h"`
	MarketCapChange24h        float64     `json:"market_cap_change_24h"`
	MarketCapChangePercent24h float64     `json:"market_cap_change_percentage_24h"`
	CirculatingSupply         float64     `json:"circulating_supply"`
	TotalSupply               *float64    `json:"total_supply"`
	MaxSupply                 *float64    `json:"max_supply"`
	ATH                       float64     `json:"ath"`
	ATHChangePercent          float64     `json:"ath_change_percentage"`
	ATHDate                   time.Time   `json:"ath_date"`
	ATL                       float64     `json:"atl"`
	ATLChangePercent          float64     `json:"atl_change_percentage"`
	ATLDate                   time.Time   `json:"atl_date"`
	ROI                       interface{} `json:"roi"`
	LastUpdated               time.Time   `json:"last_updated"`
}

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

func (c *CoinGeckoClient) GetCoins(ctx context.Context) ([]CoinGeckoCoin, error) {
	reqURL := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1&sparkline=false&locale=en", c.baseURL, c.coinLimit)
	if c.apiKey != "" {
		reqURL += "&x_cg_pro_api_key=" + c.apiKey
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
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
