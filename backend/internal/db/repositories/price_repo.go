package repositories

import (
	"context"
	"strings"
	"tracker/internal/core"
	"tracker/internal/core/domain"
	"tracker/internal/db"
	"tracker/internal/db/models"

	"github.com/jackc/pgx/v5/pgtype"
)

type PriceRepository struct {
	db *db.DataBase
}

func NewPriceRepository(db *db.DataBase) core.PriceRepository {
	return &PriceRepository{db: db}
}

func (r *PriceRepository) GetCoins(ctx context.Context) ([]domain.TokenID, error) {
	query := `SELECT id, symbol, coin_name, image_url, updated_at FROM coins`

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
			&snapshot.UpdatedAt,
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

func (r *PriceRepository) SetCoins(ctx context.Context, in []domain.TokenID) error {
	if len(in) == 0 {
		return nil
	}
	query := `INSERT INTO coins (
			id,
			symbol,
			coin_name,
			image_url,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5
		)
		ON CONFLICT (id)
		DO UPDATE SET
			symbol = EXCLUDED.symbol,
			coin_name = EXCLUDED.coin_name,
			image_url = EXCLUDED.image_url,
			updated_at = EXCLUDED.updated_at
	`
	for _, snapshot := range domainTokensToModel(in) {
		_, err := r.db.Pool.Exec(ctx, query,
			strings.ToUpper(snapshot.ID),
			strings.ToUpper(snapshot.Symbol),
			snapshot.Name,
			snapshot.ImageURL,
			snapshot.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PriceRepository) GetPrices(ctx context.Context) ([]domain.TokenPrice, error) {
	query := `SELECT 
		coins.id,
		coins.symbol,
		coins.coin_name,
		coin_prices.price_usd,
		coin_prices.market_cap_usd,
		coin_prices.total_volume_usd,
		coin_prices.price_change_24h,
		coin_prices.price_change_percent_24h,
		coin_prices.market_cap_change_24h,
		coin_prices.market_cap_change_percent_24h,
		coin_prices.updated_at
		FROM coin_prices
	JOIN coins
	ON coins.id = coin_prices.id`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.Price
	for rows.Next() {
		var snapshot models.Price
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
			&snapshot.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if snapshot.Symbol == "XRP" {
			print("hooh")
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return modelPriceToDomain(snapshots), nil
}

func (r *PriceRepository) SetPrices(ctx context.Context, in []domain.TokenPrice) error {
	if len(in) == 0 {
		return nil
	}
	query := `INSERT INTO coin_prices
		(	
			id,
			price_usd,
			market_cap_usd,
			total_volume_usd,
			price_change_24h,
			price_change_percent_24h,
			market_cap_change_24h,
			market_cap_change_percent_24h,
			updated_at
		)
		VALUES (
			(SELECT id from coins where id = $1 AND symbol = $2),
			$3, $4, $5, $6, $7, $8, $9, $10
		)
		ON CONFLICT (id)
		DO UPDATE SET
			price_usd = EXCLUDED.price_usd,
			market_cap_usd = EXCLUDED.market_cap_usd,
			total_volume_usd = EXCLUDED.total_volume_usd,
			price_change_24h = EXCLUDED.price_change_24h,
			price_change_percent_24h = EXCLUDED.price_change_percent_24h,
			market_cap_change_24h = EXCLUDED.market_cap_change_24h,
			market_cap_change_percent_24h = EXCLUDED.market_cap_change_percent_24h,
			updated_at = EXCLUDED.updated_at
		RETURNING 
			id, price_usd, 
			market_cap_usd,
			total_volume_usd,
			price_change_24h,
			price_change_percent_24h,
			market_cap_change_24h,
			market_cap_change_percent_24h,
			updated_at
	`
	for _, snapshot := range domainPriceToModel(in) {
		_, err := r.db.Pool.Exec(ctx, query,
			strings.ToUpper(snapshot.ID),
			strings.ToUpper(snapshot.Symbol),
			snapshot.CurrentPrice,
			snapshot.MarketCap,
			snapshot.TotalVolume,
			snapshot.Change_24h,
			snapshot.PriceChangePercentage_24h,
			snapshot.MarketCapChange_24h,
			snapshot.MarketCapChange_percentage_24h,
			snapshot.UpdatedAt,
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

func domainPriceToModel(in []domain.TokenPrice) []models.Price {
	out := make([]models.Price, 0, len(in))
	for _, i := range in {
		out = append(out, models.Price{
			Symbol:                         i.Symbol,
			Name:                           i.Name,
			ID:                             i.ID,
			CurrentPrice:                   pgtype.Float8{Float64: i.CurrentPrice, Valid: true},
			Change_24h:                     pgtype.Float8{Float64: i.Change_24h, Valid: true},
			MarketCap:                      pgtype.Float8{Float64: i.MarketCap, Valid: true},
			TotalVolume:                    pgtype.Float8{Float64: i.TotalVolume, Valid: true},
			High_24h:                       pgtype.Float8{Float64: i.High_24h, Valid: true},
			Low_24h:                        pgtype.Float8{Float64: i.Low_24h, Valid: true},
			PriceChange_24h:                pgtype.Float8{Float64: i.PriceChange_24h, Valid: true},
			PriceChangePercentage_24h:      pgtype.Float8{Float64: i.PriceChangePercentage_24h, Valid: true},
			MarketCapChange_24h:            pgtype.Float8{Float64: i.MarketCapChange_24h, Valid: true},
			MarketCapChange_percentage_24h: pgtype.Float8{Float64: i.MarketCapChange_percentage_24h, Valid: true},
			UpdatedAt:                      pgtype.Timestamptz{Time: i.UpdatedAt, Valid: true},
		})
	}
	return out
}

func modelPriceToDomain(in []models.Price) []domain.TokenPrice {
	out := make([]domain.TokenPrice, 0, len(in))
	for _, i := range in {
		out = append(out, domain.TokenPrice{
			ID:                             i.ID,
			Symbol:                         i.Symbol,
			Name:                           i.Name,
			CurrentPrice:                   i.CurrentPrice.Float64,
			Change_24h:                     i.Change_24h.Float64,
			MarketCap:                      i.MarketCap.Float64,
			TotalVolume:                    i.TotalVolume.Float64,
			High_24h:                       i.High_24h.Float64,
			Low_24h:                        i.Low_24h.Float64,
			PriceChange_24h:                i.PriceChange_24h.Float64,
			PriceChangePercentage_24h:      i.PriceChangePercentage_24h.Float64,
			MarketCapChange_24h:            i.MarketCapChange_24h.Float64,
			MarketCapChange_percentage_24h: i.MarketCapChange_percentage_24h.Float64,
			UpdatedAt:                      i.UpdatedAt.Time,
		})
	}
	return out
}
