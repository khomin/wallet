package dto

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `db:"id"`
	Address   string    `db:"address"`
	Chain     string    `db:"chain"`
	Label     string    `db:"label"`
	UserID    string    `db:"user_id"` // Internal only
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"` // Not needed by client
}
