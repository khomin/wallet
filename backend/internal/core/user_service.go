package core

import (
	"context"
	"tracker/internal/core/domain"
)

type UserRepo interface {
	List(ctx context.Context) ([]domain.User, error)
	EnsureExists(ctx context.Context, user *domain.User) error
}
