package core

import (
	"context"
	"tracker/internal/core/domain"
)

type AlertRepository interface {
	ListByUser(ctx context.Context, userID string) ([]domain.Alert, error)
	ListActive(ctx context.Context) ([]domain.Alert, error)

	Create(ctx context.Context, alert domain.Alert) (*domain.Alert, error)
	Update(ctx context.Context, userID, id string, alert domain.AlertUpdate) (*domain.Alert, error)
	Enable(ctx context.Context, userID, id string) (*domain.Alert, error)
	Disable(ctx context.Context, userID, id string) (*domain.Alert, error)
	Delete(ctx context.Context, userID string, id string) error
}
