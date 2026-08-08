package models

import (
	"time"
)

type Coin struct {
	ID        string
	Symbol    string
	Name      string
	ImageURL  string
	UpdatedAt time.Time
}

type Price struct {
	ID                             string
	Name                           string
	Symbol                         string
	CurrentPrice                   float64
	Change_24h                     float64
	MarketCap                      float64
	TotalVolume                    float64
	High_24h                       float64
	Low_24h                        float64
	PriceChange_24h                float64
	PriceChangePercentage_24h      float64
	MarketCapChange_24h            float64
	MarketCapChange_percentage_24h float64
	UpdatedAt                      time.Time
}
