package domain

import (
	"time"

	"github.com/google/uuid"
)

type Alert struct {
	id          uuid.UUID
	UserID      string
	coin_id     string
	condition   string
	price       float64
	Enabled     bool
	TriggeredAt *time.Time
	UpdatedAt   time.Time
	CreatedAt   time.Time
}
