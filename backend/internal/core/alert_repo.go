package core

import (
	"context"
	"tracker/internal/core/domain"
)

type AlertRepository interface {
	ListByUser(ctx context.Context, userID string) ([]domain.Alert, error)
	ListActive(ctx context.Context) ([]domain.Alert, error)
	Create(ctx context.Context, alert domain.Alert) (*domain.Alert, error)
	Delete(ctx context.Context, id string) error
}
