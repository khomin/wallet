package repositories

import (
	"context"
	"errors"
	"strings"
	"time"
	"tracker/internal/core"
	"tracker/internal/core/domain"
	"tracker/internal/db"
	"tracker/internal/db/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type walletRepository struct {
	db *db.DataBase
}

func NewWalletRepository(db *db.DataBase) core.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) List(ctx context.Context, userID string) ([]domain.WalletBalance, error) {
	query := `
		SELECT
			w.id, w.user_id,  w.address, w.chain, coin.symbol, w.label, w.notify, w.updated_at,
			balance.value_crypto,
			balance.value_usd,
			balance.updated_at,
			coin.id,
			coin.symbol,
			coin.coin_name,
			price.price_usd,
			price.market_cap_usd,
			price.total_volume_usd,
			price.price_change_24h,
			price.price_change_percent_24h,
			price.market_cap_change_24h,
			price.market_cap_change_percent_24h,
			price.updated_at
		FROM wallets w
		LEFT JOIN coins coin ON coin.id = w.coin_id
		LEFT JOIN coin_prices price ON price.id = w.coin_id
		LEFT JOIN wallet_balances balance ON balance.id = w.id
		WHERE w.user_id = $1
		ORDER BY w.updated_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []domain.WalletBalance
	for rows.Next() {
		w, err := scanWalletBalance(rows)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, walletToDomainBalance(w))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *walletRepository) Get(ctx context.Context, userID string, id uuid.UUID) (*domain.WalletBalance, error) {
	query := `
		SELECT
			w.id, w.user_id,  w.address, w.chain, coin.symbol, w.label, w.notify, w.updated_at,
			balance.value_crypto, balance.value_usd, balance.updated_at,
			coin.id, coin.symbol, coin.coin_name,
			price.price_usd,
			price.market_cap_usd,
			price.total_volume_usd,
			price.price_change_24h,
			price.price_change_percent_24h,
			price.market_cap_change_24h,
			price.market_cap_change_percent_24h,
			price.updated_at
		FROM wallets w
		LEFT JOIN coins coin ON coin.id = w.coin_id
		LEFT JOIN coin_prices price ON price.id = w.coin_id
		LEFT JOIN wallet_balances balance ON balance.id = w.id
		WHERE w.user_id = $1 AND w.id = $2
	`
	rows, err := r.db.Pool.Query(ctx, query,
		userID,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		w, err := scanWalletBalance(rows)
		if err != nil {
			return nil, err
		}
		out := walletToDomainBalance(w)
		return &out, nil
	}
	return nil, domain.ErrorNotFound
}

func (r *walletRepository) Create(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.Wallet, error) {
	query := `
		INSERT INTO wallets (address, chain, coin_id, label, user_id)
		VALUES (
			$1, 
			$2,
			(SELECT id FROM coins WHERE symbol = $3), 
			$4, $5
		)
		RETURNING id, user_id, address, chain, (SELECT symbol FROM coins WHERE symbol = $3), label, updated_at
	`
	row := r.db.Pool.QueryRow(ctx, query,
		address,
		strings.ToUpper(chain),
		strings.ToUpper(symbol),
		label,
		userID,
	)
	var wallet models.Wallet
	if err := row.Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Address,
		&wallet.Chain,
		&wallet.Symbol,
		&wallet.Label,
		&wallet.UpdatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrorWWalletAlreadyExists
		}
		return nil, domain.ErrorWWalletInternalError
	}
	out := walletToDomain(wallet)
	return &out, nil
}

func (r *walletRepository) Update(ctx context.Context, userID string, id uuid.UUID, req core.UpdateWallet) (*domain.Wallet, error) {
	query := `
		UPDATE wallets
		SET 
			label = $3,
			notify = $4
			WHERE id = $1 AND user_id = $2
		RETURNING 
			id, user_id, address, chain, 
			(SELECT coins.symbol FROM coins WHERE coins.id = wallets.coin_id),
			label, notify, updated_at
	`
	row := r.db.Pool.QueryRow(ctx, query,
		id,
		userID,
		req.Label,
		req.Notify,
	)
	var wallet models.Wallet
	if err := row.Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Address,
		&wallet.Chain,
		&wallet.Symbol,
		&wallet.Label,
		&wallet.Notify,
		&wallet.UpdatedAt,
	); err != nil {
		return nil, domain.ErrorNotFound
	}
	out := walletToDomain(wallet)
	return &out, nil
}

func (r *walletRepository) Delete(ctx context.Context, userID string, id uuid.UUID) error {
	query := `DELETE FROM wallets WHERE id = $1`
	res, err := r.db.Pool.Exec(ctx, query, id)
	if res.RowsAffected() == 0 {
		return domain.ErrorNotFound
	}
	return err
}

func (r *walletRepository) UpdateBalanceSnapshot(ctx context.Context, userID string, id uuid.UUID, snapshot core.BalanceSnapshot) error {
	query := `
		WITH updated AS (
			INSERT INTO wallet_balances (
				id,
				value_crypto,
				value_usd,
				updated_at
			)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE
			SET
				value_crypto = EXCLUDED.value_crypto,
				value_usd = EXCLUDED.value_usd,
				updated_at = EXCLUDED.updated_at
			RETURNING
				id,
				value_crypto,
				value_usd
		)
		INSERT INTO wallet_balance_snapshots (
			wallet_id,
			value_crypto,
			value_usd,
			created_at
		)
		SELECT
			u.id,
			u.value_crypto,
			u.value_usd,
			$4
		FROM updated u
		WHERE NOT EXISTS (
			SELECT 1
			FROM wallet_balance_snapshots s
			WHERE s.wallet_id = u.id
			AND s.value_crypto IS NOT DISTINCT FROM u.value_crypto
			AND s.value_usd IS NOT DISTINCT FROM u.value_usd
			ORDER BY s.created_at DESC
			LIMIT 1
		);
	`
	_, err := r.db.Pool.Exec(ctx, query,
		id,
		snapshot.Crypto,
		snapshot.USD,
		snapshot.Time,
	)
	if err != nil {
		return err
	}
	return err
}

func (r *walletRepository) GetBalanceSnapshot(ctx context.Context, userID string, id uuid.UUID, filter core.BalanceSnapshotFilter) ([]domain.WalletBalanceSnapshot, error) {
	if filter.Limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}
	query := `
		WITH snapshots AS (
			SELECT
				wb.value_crypto,
				wb.value_usd,
				wb.created_at,
				row_number() OVER (ORDER BY wb.created_at) AS rn,
				count(*) OVER () AS total
			FROM wallet_balance_snapshots wb
			JOIN wallets w ON w.id = wb.wallet_id
			WHERE wb.wallet_id = $1
			AND w.user_id = $2
			AND wb.created_at >= $3
			AND wb.created_at <= $4
		),
		positions AS (
			SELECT DISTINCT
				ROUND(
					i * (total - 1)::numeric / ($5 - 1)
				) + 1 AS rn
			FROM snapshots
			CROSS JOIN generate_series(0, $5 - 1) AS i
		)
		SELECT
			s.value_crypto,
			s.value_usd,
			s.created_at
		FROM snapshots s
		JOIN positions p ON p.rn = s.rn
		ORDER BY s.created_at;
	`
	rows, err := r.db.Pool.Query(ctx, query,
		id,
		userID,
		filter.From,
		filter.To,
		filter.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var out []domain.WalletBalanceSnapshot
	for rows.Next() {
		var i domain.WalletBalanceSnapshot
		err := rows.Scan(
			&i.Balance,
			&i.BalanceUSD,
			&i.Time,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, nil
}

func (r *walletRepository) ListForSync(ctx context.Context, updatedAt time.Time, limit int) ([]domain.Wallet, error) {
	query := `
		SELECT 
			wallets.id, 
			wallets.user_id, 
			wallets.address, 
			wallets.chain, 
			coins.symbol,
			wallets.label,
			wallets.updated_at 
		FROM wallets
		LEFT JOIN coins 
			ON coins.id = wallets.coin_id
		LEFT JOIN wallet_balances balance
			ON balance.id = wallets.id
		WHERE balance.updated_at IS NULL OR balance.updated_at < $1
		ORDER BY balance.updated_at ASC NULLS FIRST
		LIMIT $2
	`
	rows, err := r.db.Pool.Query(ctx, query,
		updatedAt,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Wallet{}
	for rows.Next() {
		var wallet models.Wallet
		err = rows.Scan(
			&wallet.ID,
			&wallet.UserID,
			&wallet.Address,
			&wallet.Chain,
			&wallet.Symbol,
			&wallet.Label,
			&wallet.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, walletToDomain(wallet))
	}
	return out, nil
}

func scanWalletBalance(rows pgx.Rows) (*models.WalletBalance, error) {
	var i models.WalletBalance
	err := rows.Scan(
		&i.Wallet.ID,
		&i.Wallet.UserID,
		&i.Wallet.Address,
		&i.Wallet.Chain,
		&i.Wallet.Symbol,
		&i.Wallet.Label,
		&i.Wallet.Notify,
		&i.Wallet.UpdatedAt,
		//
		&i.Balance,
		&i.BalanceUSD,
		&i.UpdatedAt,
		//
		&i.Price.ID,
		&i.Price.Symbol,
		&i.Price.Name,
		&i.Price.CurrentPrice,
		&i.Price.MarketCap,
		&i.Price.TotalVolume,
		&i.Price.Change_24h,
		&i.Price.PriceChangePercentage_24h,
		&i.Price.MarketCapChange_24h,
		&i.Price.MarketCapChange_percentage_24h,
		&i.Price.UpdatedAt,
	)
	return &i, err
}

func walletToDomain(in models.Wallet) domain.Wallet {
	return domain.Wallet{
		ID:      in.ID.String(),
		UserID:  in.UserID,
		Address: in.Address,
		Chain:   in.Chain,
		Symbol:  in.Symbol,
		Label:   in.Label,
		Notify:  in.Notify,
	}
}

func walletToDomainBalance(in *models.WalletBalance) domain.WalletBalance {
	hasError := false
	var errorMsg string
	if !in.BalanceUSD.Valid || !in.Balance.Valid {
		hasError = true
		errorMsg = "Unable to fetch live balance"
	}
	return domain.WalletBalance{
		Wallet: domain.Wallet{
			ID:      in.Wallet.ID.String(),
			UserID:  in.Wallet.UserID,
			Address: in.Wallet.Address,
			Chain:   in.Wallet.Chain,
			Symbol:  in.Wallet.Symbol,
			Label:   in.Wallet.Label,
			Notify:  in.Wallet.Notify,
		},
		Price: domain.TokenPrice{
			ID:                             in.Price.ID,
			Name:                           in.Price.Name,
			Symbol:                         in.Price.Symbol,
			CurrentPrice:                   in.Price.CurrentPrice.Float64,
			Change_24h:                     in.Price.Change_24h.Float64,
			MarketCap:                      in.Price.MarketCap.Float64,
			TotalVolume:                    in.Price.TotalVolume.Float64,
			High_24h:                       in.Price.High_24h.Float64,
			Low_24h:                        in.Price.Low_24h.Float64,
			PriceChange_24h:                in.Price.PriceChange_24h.Float64,
			PriceChangePercentage_24h:      in.Price.PriceChangePercentage_24h.Float64,
			MarketCapChange_24h:            in.Price.MarketCapChange_24h.Float64,
			MarketCapChange_percentage_24h: in.Price.MarketCapChange_percentage_24h.Float64,
			UpdatedAt:                      in.Price.UpdatedAt.Time,
		},
		Balance:    in.Balance.Float64,
		BalanceUSD: in.BalanceUSD.Float64,
		UpdatedAt:  in.UpdatedAt.Time,
		HasError:   hasError,
		ErrorMsg:   errorMsg,
	}
}
