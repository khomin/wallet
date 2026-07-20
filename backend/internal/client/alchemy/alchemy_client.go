package alchemy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tracker/internal/db/models"

	"github.com/sirupsen/logrus"
)

type AlchemyClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewAlchemyClient(apiKey string) *AlchemyClient {
	return &AlchemyClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
		baseURL:    "https://api.g.alchemy.com/prices/v1",
	}
}

func (c *AlchemyClient) GetPrices(ctx context.Context, symbols []string) ([]models.CoinPrice, error) {
	if len(symbols) > 25 {
		return nil, fmt.Errorf("maximum 25 symbols per request")
	}
	str := strings.Join(symbols, ",")
	req := fmt.Sprintf("%s/%s/tokens/by-symbol?symbols=%s", c.baseURL, c.apiKey, str)
	resp, err := http.Get(req)
	if err != nil {
		return nil, err
	}
	var prices []models.CoinPrice
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var result AlchemyPriceResponse
		if err := json.Unmarshal(body, &result); err != nil {
			logrus.Errorf("Can't decode price: %s [%s]", string(body), err.Error())
		}
		// Transform to our internal format
		for _, token := range result.Data {
			if token.Error != nil {
				continue
			}
			for _, p := range token.Prices {
				if strings.ToLower(p.Currency) == "usd" {
					price, error := strconv.ParseFloat(p.Value, 2)
					if error == nil {
						prices = append(prices, models.CoinPrice{
							Symbol:       token.Symbol,
							CoinID:       "",
							Name:         "",
							CurrentPrice: price,
							LastUpdated:  p.LastUpdated,
						})
					}
				}
			}
		}
		return prices, nil
	}
	return prices, nil
}
