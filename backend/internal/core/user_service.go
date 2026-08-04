package core

import "context"

type UserRepo interface {
	EnsureExists(ctx context.Context, userID string) error
}
