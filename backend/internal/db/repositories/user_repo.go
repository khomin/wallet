package repositories

import (
	"context"
	"tracker/internal/core"
	"tracker/internal/core/domain"
	"tracker/internal/db"
)

type userRepo struct {
	db *db.DataBase
}

func NewUserRepo(db *db.DataBase) core.UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) List(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, name, email FROM users`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.User{}
	for rows.Next() {
		var i domain.User
		err = rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, nil
}

func (r *userRepo) EnsureExists(ctx context.Context, user domain.User) error {
	query := `INSERT INTO users (id, name, email) 
		VALUES ($1, $2, $3)
		ON CONFLICT (id)
		DO NOTHING`
	_, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
	)
	return err
}
