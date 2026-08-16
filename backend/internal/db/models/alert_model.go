package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Alert struct {
	ID          pgtype.UUID
	UserID      string
	CoinID      string
	Condition   string
	Price       pgtype.Float8
	Enabled     bool
	TriggeredAt pgtype.Timestamptz
	CreatedAt   time.Time
	UpdatedAt   time.Time
}