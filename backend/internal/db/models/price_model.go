package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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
	CurrentPrice                   pgtype.Float8
	Change_24h                     pgtype.Float8
	MarketCap                      pgtype.Float8
	TotalVolume                    pgtype.Float8
	High_24h                       pgtype.Float8
	Low_24h                        pgtype.Float8
	PriceChange_24h                pgtype.Float8
	PriceChangePercentage_24h      pgtype.Float8
	MarketCapChange_24h            pgtype.Float8
	MarketCapChange_percentage_24h pgtype.Float8
	UpdatedAt                      pgtype.Timestamptz
}
