package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Address   string    `db:"address" json:"address"`
	Chain     string    `db:"chain" json:"chain"`
	Label     string    `db:"label" json:"label"`
	UserID    string    `db:"user_id" json:"user_id,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
