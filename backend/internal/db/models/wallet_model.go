package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `db:"id"`
	Address   string    `db:"address" `
	Chain     string    `db:"chain" `
	Label     string    `db:"label"`
	Symbol    string    `db:"symbol"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at" `
	UpdatedAt time.Time `db:"updated_at" `
}
