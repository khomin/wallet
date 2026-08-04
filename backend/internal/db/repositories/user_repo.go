package repositories

import (
	"context"
	"tracker/internal/core"
	"tracker/internal/db"
)

type userRepo struct {
	db *db.DataBase
}

func NewUserRepo(db *db.DataBase) core.UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) EnsureExists(ctx context.Context, userID string) error {
	query := `INSERT INTO users (id) VALUES ($1)
			ON CONFLICT (id)
			DO NOTHING`
	_, err := r.db.Pool.Exec(ctx, query, userID)
	return err
}
