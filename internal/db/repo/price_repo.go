package repositories

import (
	"context"
	"tracker/internal/db"
	"tracker/internal/db/models"
)

type PriceRepository struct {
	db *db.DataBase
}

func NewPriceRepository(db *db.DataBase) *PriceRepository {
	return &PriceRepository{db: db}
}

func (r *PriceRepository) GetCoinSnapshot(ctx context.Context) ([]models.CoinSnapshot, error) {
	query := `SELECT
		id,
		coin_id,
		symbol,
		coin_name,
		image_url,
		last_updated,
		snapshot_at
	FROM coin_snapshots
	ORDER BY coin_id ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.CoinSnapshot
	for rows.Next() {
		var snapshot models.CoinSnapshot
		if err := rows.Scan(
			&snapshot.ID,
			&snapshot.CoinID,
			&snapshot.Symbol,
			&snapshot.Name,
			&snapshot.ImageURL,
			&snapshot.LastUpdated,
			&snapshot.SnapshotAt,
		); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (r *PriceRepository) SetCoinSnapshot(ctx context.Context, snapshots []models.CoinSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	query := `INSERT INTO coin_snapshots (
			coin_id,
			symbol,
			coin_name,
			image_url,
			last_updated,
			snapshot_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6
		)
		ON CONFLICT (coin_id)
		DO UPDATE SET
			symbol = EXCLUDED.symbol,
			coin_name = EXCLUDED.coin_name,
			image_url = EXCLUDED.image_url,
			last_updated = EXCLUDED.last_updated,
			snapshot_at = EXCLUDED.snapshot_at
	`
	for _, snapshot := range snapshots {
		_, err := r.db.Pool.Exec(ctx, query,
			snapshot.CoinID,
			snapshot.Symbol,
			snapshot.Name,
			snapshot.ImageURL,
			snapshot.LastUpdated,
			snapshot.SnapshotAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PriceRepository) GetPriceSnapshot(ctx context.Context) ([]models.PriceSnapshot, error) {
	query := `SELECT
		id,
		coin_id,
		symbol,
		coin_name,
		image_url,
		last_updated,
		snapshot_at
	FROM coin_snapshots
	ORDER BY coin_id ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.PriceSnapshot
	for rows.Next() {
		var snapshot models.PriceSnapshot
		if err := rows.Scan(
			&snapshot.ID,
			&snapshot.Symbol,
			&snapshot.PriceUSD,
			&snapshot.Change24h,
			&snapshot.LastUpdated,
		); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (r *PriceRepository) SetPriceSnapshot(ctx context.Context, snapshots []models.PriceSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	query := `INSERT INTO price_snapshots (
			coin_id,
			symbol,
			price_usd,
			change_24h,
			last_updated,
		)
		VALUES (
			$1, $2, $3, $4, $5, $6
		)
		ON CONFLICT (coin_id)
		DO UPDATE SET
			symbol = EXCLUDED.symbol,
			coin_name = EXCLUDED.coin_name,
			image_url = EXCLUDED.image_url,
			last_updated = EXCLUDED.last_updated,
			snapshot_at = EXCLUDED.snapshot_at
	`
	for _, snapshot := range snapshots {
		_, err := r.db.Pool.Exec(ctx, query,
			snapshot.ID,
			snapshot.Symbol,
			snapshot.PriceUSD,
			snapshot.Change24h,
			snapshot.LastUpdated,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
