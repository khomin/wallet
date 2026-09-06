package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Wallet struct {
	ID        pgtype.UUID
	Address   string
	Chain     string
	Label     string
	Notify    bool
	Symbol    string
	UserID    string
	UpdatedAt pgtype.Timestamptz
}

type WalletBalance struct {
	Wallet     Wallet
	Price      Price
	Balance    pgtype.Float8
	BalanceUSD pgtype.Float8
	UpdatedAt  pgtype.Timestamptz
	HasError   bool
	ErrorMsg   string
}
