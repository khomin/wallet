package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DataBase struct {
	Dsn    string
	Config *pgxpool.Config
	Pool   *pgxpool.Pool
}

func NewDatabase(dsn string) (*DataBase, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &DataBase{
		Dsn:    dsn,
		Config: config,
		Pool:   pool,
	}, nil
}

func (d *DataBase) Close() {
	if d.Pool != nil {
		defer d.Pool.Close()
	}
}
