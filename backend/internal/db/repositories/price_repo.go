package repositories

import (
	"context"
	"time"
	"tracker/internal/db"
	"tracker/internal/db/models"
)

type PriceRepository struct {
	db *db.DataBase
}

func NewPriceRepository(db *db.DataBase) PriceRepository {
	return PriceRepository{db: db}
}

func (r *PriceRepository) GetCoinSnapshot(ctx context.Context) ([]models.Coin, error) {
	query := `SELECT
		id,
		coin_id,
		symbol,
		coin_name,
		image_url,
		last_updated,
		snapshot_at
	FROM coins
	ORDER BY coin_id ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.Coin
	for rows.Next() {
		var snapshot models.Coin
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

func (r *PriceRepository) SetCoinSnapshot(ctx context.Context, snapshots []models.Coin) error {
	if len(snapshots) == 0 {
		return nil
	}
	query := `INSERT INTO coins (
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

func (r *PriceRepository) GetPriceSnapshot(ctx context.Context) ([]models.CoinPrice, error) {
	query := `SELECT
		coin_id,
		symbol,
		coin_name,
		price_usd,
		market_cap_usd,
		total_volume_usd,
		price_change_24h,
		price_change_percent_24h,
		market_cap_change_24h,
		market_cap_change_percent_24h,
		last_updated
	FROM coin_price_snapshots
	ORDER BY coin_id ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.CoinPrice
	for rows.Next() {
		var snapshot models.CoinPrice
		if err := rows.Scan(
			&snapshot.CoinID,
			&snapshot.Symbol,
			&snapshot.Name,
			&snapshot.CurrentPrice,
			&snapshot.MarketCap,
			&snapshot.TotalVolume,
			&snapshot.Change_24h,
			&snapshot.PriceChangePercentage_24h,
			&snapshot.MarketCapChange_24h,
			&snapshot.MarketCapChange_percentage_24h,
			&snapshot.LastUpdated,
		); err != nil {
			return nil, err
		}
		snapshot.PriceChange_24h = snapshot.Change_24h
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (r *PriceRepository) SetPriceSnapshot(ctx context.Context, snapshots []models.CoinPrice) error {
	if len(snapshots) == 0 {
		return nil
	}
	query := `INSERT INTO coin_price_snapshots (
			coin_id,
			symbol,
			coin_name,
			price_usd,
			market_cap_usd,
			total_volume_usd,
			price_change_24h,
			price_change_percent_24h,
			market_cap_change_24h,
			market_cap_change_percent_24h,
			last_updated,
			snapshot_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		ON CONFLICT (coin_id)
		DO UPDATE SET
			symbol = EXCLUDED.symbol,
			coin_name = EXCLUDED.coin_name,
			price_usd = EXCLUDED.price_usd,
			market_cap_usd = EXCLUDED.market_cap_usd,
			total_volume_usd = EXCLUDED.total_volume_usd,
			price_change_24h = EXCLUDED.price_change_24h,
			price_change_percent_24h = EXCLUDED.price_change_percent_24h,
			market_cap_change_24h = EXCLUDED.market_cap_change_24h,
			market_cap_change_percent_24h = EXCLUDED.market_cap_change_percent_24h,
			last_updated = EXCLUDED.last_updated,
			snapshot_at = EXCLUDED.snapshot_at
	`
	for _, snapshot := range snapshots {
		_, err := r.db.Pool.Exec(ctx, query,
			snapshot.CoinID,
			snapshot.Symbol,
			snapshot.Name,
			snapshot.CurrentPrice,
			snapshot.MarketCap,
			snapshot.TotalVolume,
			snapshot.Change_24h,
			snapshot.PriceChangePercentage_24h,
			snapshot.MarketCapChange_24h,
			snapshot.MarketCapChange_percentage_24h,
			snapshot.LastUpdated,
			time.Now().UTC(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
