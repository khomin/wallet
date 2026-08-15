package repositories

import (
	"context"
	"tracker/internal/core"
	"tracker/internal/core/domain"
	"tracker/internal/db"
)

type AlertRepository struct {
	db *db.DataBase
}

func NewAlertRepository(db *db.DataBase) core.AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) Create(ctx context.Context, alert domain.Alert) (*domain.Alert, error) {
	panic("unimplemented")
}

func (r *AlertRepository) Delete(ctx context.Context, id string) error {
	panic("unimplemented")
}

func (r *AlertRepository) ListActive(ctx context.Context) ([]domain.Alert, error) {
	panic("unimplemented")
}

func (r *AlertRepository) ListByUser(ctx context.Context, userID string) ([]domain.Alert, error) {
	panic("unimplemented")
}
