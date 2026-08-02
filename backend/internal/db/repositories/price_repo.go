package repositories

import (
	"context"
	"strings"
	"time"
	"tracker/internal/core/domain"
	"tracker/internal/db"
	"tracker/internal/db/models"
)

type PriceRepository struct {
	db *db.DataBase
}

func NewPriceRepository(db *db.DataBase) PriceRepository {
	return PriceRepository{db: db}
}

func (r *PriceRepository) GetCoinSnapshot(ctx context.Context) ([]domain.TokenID, error) {
	query := `SELECT * FROM coins`

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
	return modelTokensToDomain(snapshots), nil
}

func (r *PriceRepository) SetCoinSnapshot(ctx context.Context, in []domain.TokenID) error {
	if len(in) == 0 {
		return nil
	}
	query := `INSERT INTO coins (
			id,
			symbol,
			coin_name,
			image_url,
			last_updated,
			snapshot_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6
		)
		ON CONFLICT (id)
		DO UPDATE SET
			symbol = EXCLUDED.symbol,
			coin_name = EXCLUDED.coin_name,
			image_url = EXCLUDED.image_url,
			last_updated = EXCLUDED.last_updated,
			snapshot_at = EXCLUDED.snapshot_at
	`
	for _, snapshot := range domainTokensToModel(in) {
		_, err := r.db.Pool.Exec(ctx, query,
			strings.ToUpper(snapshot.ID),
			strings.ToUpper(snapshot.Symbol),
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

func (r *PriceRepository) GetPriceSnapshot(ctx context.Context) ([]domain.TokenPrice, error) {
	query := `SELECT
		id,
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
	FROM coin_price_snapshots`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.CoinPrice
	for rows.Next() {
		var snapshot models.CoinPrice
		if err := rows.Scan(
			&snapshot.ID,
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
	return modelPriceToDomain(snapshots), nil
}

func (r *PriceRepository) SetPriceSnapshot(ctx context.Context, in []domain.TokenPrice) error {
	if len(in) == 0 {
		return nil
	}
	query := `INSERT INTO coin_price_snapshots (
			id,
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
		ON CONFLICT (id)
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
	for _, snapshot := range domainPriceToModel(in) {
		_, err := r.db.Pool.Exec(ctx, query,
			strings.ToUpper(snapshot.ID),
			strings.ToUpper(snapshot.Symbol),
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

func modelTokensToDomain(in []models.Coin) []domain.TokenID {
	out := make([]domain.TokenID, 0, len(in))
	for _, i := range in {
		out = append(out, domain.TokenID{
			ID:       i.ID,
			Symbol:   i.Symbol,
			Name:     i.Name,
			ImageURL: i.ImageURL,
		})
	}
	return out
}

func domainTokensToModel(in []domain.TokenID) []models.Coin {
	out := make([]models.Coin, 0, len(in))
	for _, i := range in {
		out = append(out, models.Coin{
			Symbol:   i.Symbol,
			Name:     i.Name,
			ID:       i.ID,
			ImageURL: i.ImageURL,
		})
	}
	return out
}

func domainPriceToModel(in []domain.TokenPrice) []models.CoinPrice {
	out := make([]models.CoinPrice, 0, len(in))
	for _, i := range in {
		out = append(out, models.CoinPrice{
			Symbol: i.Symbol,
			Name:   i.Name,
			ID:     i.ID,
		})
	}
	return out
}

func modelPriceToDomain(in []models.CoinPrice) []domain.TokenPrice {
	out := make([]domain.TokenPrice, 0, len(in))
	for _, i := range in {
		out = append(out, domain.TokenPrice{
			Symbol: i.Symbol,
			Name:   i.Name,
			// Chains: i.,
			// Address: i.,
		})
	}
	return out
}
