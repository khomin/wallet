package repositories

import (
	"context"
	"tracker/config"
	"tracker/internal/db"
)

func prepare() (context.Context, *db.DataBase, error) {
	ctx := context.Background()
	cfg := config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "tracker_admin",
			Password: "super_secure_password",
			Name:     "whale_tracker",
		},
	}
	db, err := db.NewDatabase(cfg.DSN())
	return ctx, db, err
}
