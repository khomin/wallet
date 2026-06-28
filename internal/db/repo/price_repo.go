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

func (r *PriceRepository) SetCoinSnapshot(ctx context.Context, snapshots []models.CoinSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	query := `INSERT INTO coin_snapshots (
			coin_id,
			symbol,
			coin_name,
			price_usd,
			market_cap_usd,
			market_cap_rank,
			total_volume_usd,
			price_change_24h,
			price_change_percent_24h,
			market_cap_change_24h,
			market_cap_change_percent_24h,
			circulating_supply,
			total_supply,
			max_supply,
			ath,
			ath_change_percent,
			ath_date,
			atl,
			atl_change_percent,
			atl_date,
			image_url,
			last_updated,
			snapshot_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23
		)
		ON CONFLICT (coin_id)
		DO UPDATE SET
			symbol = EXCLUDED.symbol,
			coin_name = EXCLUDED.coin_name,
			price_usd = EXCLUDED.price_usd,
			market_cap_usd = EXCLUDED.market_cap_usd,
			market_cap_rank = EXCLUDED.market_cap_rank,
			total_volume_usd = EXCLUDED.total_volume_usd,
			price_change_24h = EXCLUDED.price_change_24h,
			price_change_percent_24h = EXCLUDED.price_change_percent_24h,
			market_cap_change_24h = EXCLUDED.market_cap_change_24h,
			market_cap_change_percent_24h = EXCLUDED.market_cap_change_percent_24h,
			circulating_supply = EXCLUDED.circulating_supply,
			total_supply = EXCLUDED.total_supply,
			max_supply = EXCLUDED.max_supply,
			ath = EXCLUDED.ath,
			ath_change_percent = EXCLUDED.ath_change_percent,
			ath_date = EXCLUDED.ath_date,
			atl = EXCLUDED.atl,
			atl_change_percent = EXCLUDED.atl_change_percent,
			atl_date = EXCLUDED.atl_date,
			image_url = EXCLUDED.image_url,
			last_updated = EXCLUDED.last_updated,
			snapshot_at = EXCLUDED.snapshot_at
	`
	for _, snapshot := range snapshots {
		_, err := r.db.Pool.Exec(ctx, query,
			snapshot.CoinID,
			snapshot.Symbol,
			snapshot.Name,
			snapshot.PriceUSD,
			snapshot.MarketCapUSD,
			snapshot.MarketCapRank,
			snapshot.TotalVolumeUSD,
			snapshot.PriceChange24h,
			snapshot.PriceChangePercent24h,
			snapshot.MarketCapChange24h,
			snapshot.MarketCapChangePercent24h,
			snapshot.CirculatingSupply,
			snapshot.TotalSupply,
			snapshot.MaxSupply,
			snapshot.ATH,
			snapshot.ATHChangePercent,
			snapshot.ATHDate,
			snapshot.ATL,
			snapshot.ATLChangePercent,
			snapshot.ATLDate,
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
