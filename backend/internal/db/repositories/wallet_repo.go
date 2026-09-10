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

func (r *walletRepository) ListUsersByWallet(ctx context.Context, id uuid.UUID) ([]domain.UserWallet, error) {
	query := `
		SELECT
			w.id,
			uw.user_id, 
			w.address,
			w.chain,
			coin.symbol,
			uw.label,
			uw.notify
		FROM wallets w
		LEFT JOIN user_wallets uw on uw.wallet_id = w.id
		LEFT JOIN coins coin ON coin.id = w.coin_id
		LEFT JOIN wallet_balances balance ON balance.id = w.id
		WHERE w.id = $1
		ORDER BY w.updated_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.UserWallet
	for rows.Next() {
		w, err := scanUserWallet(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *walletRepository) GetWalletByUser(ctx context.Context, userID string, id uuid.UUID) (*domain.UserWalletBalance, error) {
	query := `
		SELECT
			w.id, 
			w.address, 
			w.chain, 
			coin.symbol, 
			uw.label, 
			uw.notify, 
			uw.user_id,  
			w.updated_at,
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
		JOIN user_wallets uw ON uw.wallet_id = w.id
		WHERE uw.user_id = $1
		AND w.id = $2
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

func (r *walletRepository) ListWalletsByUser(ctx context.Context, userID string) ([]domain.UserWalletBalance, error) {
	query := `
		SELECT
			w.id,
			w.address,
			w.chain,
			coin.symbol,
			uw.label,
			uw.notify,
			uw.user_id, 
			w.updated_at,
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
		LEFT JOIN user_wallets uw on uw.wallet_id = w.id
		LEFT JOIN coins coin ON coin.id = w.coin_id
		LEFT JOIN coin_prices price ON price.id = w.coin_id
		LEFT JOIN wallet_balances balance ON balance.id = w.id
		WHERE uw.user_id = $1
		ORDER BY w.updated_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []domain.UserWalletBalance
	for rows.Next() {
		w, err := scanWalletBalance(rows)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrorNotFound
			}
			return nil, err
		}
		wallets = append(wallets, walletToDomainBalance(w))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *walletRepository) Create(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.UserWallet, error) {
	query := `
		WITH wallet_result AS (
			INSERT INTO wallets (address, chain, coin_id)
			VALUES (
				$1, 
				$2,
				(SELECT id FROM coins WHERE symbol = $3)
			)
			RETURNING id, address, chain, (SELECT symbol FROM coins WHERE symbol = $3), updated_at
		),
		uw_result AS (
			INSERT INTO user_wallets (user_id, wallet_id, label, notify)
			SELECT $5, wallet_result.id, $4, TRUE 
			FROM wallet_result
			RETURNING user_id, wallet_id, label, notify
		)
		SELECT id, user_id, address, chain, (SELECT symbol FROM coins WHERE symbol = $3), label, notify, updated_at from wallet_result wr
		LEFT JOIN uw_result uw ON uw.wallet_id = wr.id
		ORDER BY wr.updated_at ASC
	`
	row := r.db.Pool.QueryRow(ctx, query,
		address,
		strings.ToUpper(chain),
		strings.ToUpper(symbol),
		label,
		userID,
	)
	var wallet models.Wallet
	err := row.Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Address,
		&wallet.Chain,
		&wallet.Symbol,
		&wallet.Label,
		&wallet.Notify,
		&wallet.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrorWalletAlreadyExists
		}
		return nil, domain.ErrorInternalError
	}
	out := userWalletToDomain(wallet)
	return &out, nil
}

func (r *walletRepository) Update(ctx context.Context, userID string, id uuid.UUID, req core.UpdateWallet) (*domain.UserWallet, error) {
	query := `
		WITH wallet_update AS (
			UPDATE user_wallets
			SET 
				label = $3,
				notify = $4
				WHERE wallet_id = $1 AND user_id = $2
			RETURNING *
		)
		SELECT
			uw.wallet_id,
			uw.user_id,
			w.address,
			w.chain,
			(SELECT coins.symbol FROM coins WHERE coins.id = w.coin_id),
			uw.label, 
			uw.notify, 
			updated_at
		FROM wallet_update uw
		JOIN wallets w ON w.id = uw.wallet_id
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
	out := userWalletToDomain(wallet)
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

func (r *walletRepository) UpdateBalanceSnapshot(ctx context.Context, id uuid.UUID, snapshot core.BalanceSnapshot) error {
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
		WHERE (
			SELECT ROW(s.value_crypto, s.value_usd)
			FROM wallet_balance_snapshots s
			WHERE s.wallet_id = u.id
			ORDER BY s.created_at DESC
			LIMIT 1
		) IS DISTINCT FROM ROW(u.value_crypto, u.value_usd);
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

func (r *walletRepository) ListBalanceSnapshots(ctx context.Context, id uuid.UUID, filter core.BalanceSnapshotFilter) ([]domain.WalletBalanceSnapshot, error) {
	if filter.Limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}
	query := `
		SELECT
			wb.value_crypto,
			wb.value_usd,
			wb.created_at
		FROM wallet_balance_snapshots wb
		JOIN wallets w ON w.id = wb.wallet_id
		WHERE wb.wallet_id = $1
		AND wb.created_at >= $2
		AND wb.created_at <= $3
	`
	rows, err := r.db.Pool.Query(ctx, query,
		id,
		filter.From,
		filter.To,
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
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrorNotFound
			}
			return nil, err
		}
		out = append(out, i)
	}
	return out, nil
}

func (r *walletRepository) GetBalanceSnapshot(ctx context.Context, id uuid.UUID) (*domain.WalletBalanceSnapshot, error) {
	query := `
		SELECT
			wb.value_crypto,
			wb.value_usd,
			wb.created_at
		FROM wallet_balance_snapshots wb
		JOIN wallets w ON w.id = wb.wallet_id
		WHERE wb.wallet_id = $1
		ORDER BY wb.created_at
		LIMIT 1
	`
	row := r.db.Pool.QueryRow(ctx, query, id)
	var out domain.WalletBalanceSnapshot
	err := row.Scan(
		&out.Balance,
		&out.BalanceUSD,
		&out.Time,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorNotFound
		}
		return nil, err
	}
	return &out, nil
}

func (r *walletRepository) ListForSync(ctx context.Context, updatedAt time.Time, limit int) ([]domain.Wallet, error) {
	query := `
		SELECT 
			wallets.id, 
			wallets.address, 
			wallets.chain, 
			coins.symbol,
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
			&wallet.Address,
			&wallet.Chain,
			&wallet.Symbol,
			&wallet.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrorNotFound
			}
			return nil, err
		}
		out = append(out, walletToDomain(wallet))
	}
	return out, nil
}

func scanUserWallet(rows pgx.Rows) (*domain.UserWallet, error) {
	var i domain.UserWallet
	err := rows.Scan(
		&i.Wallet.ID,
		&i.UserID,
		&i.Wallet.Address,
		&i.Wallet.Chain,
		&i.Wallet.Symbol,
		&i.Label,
		&i.Notify,
	)
	return &i, err
}

func scanWalletBalance(rows pgx.Rows) (*models.WalletBalance, error) {
	var i models.WalletBalance
	err := rows.Scan(
		&i.Wallet.ID,
		&i.Wallet.Address,
		&i.Wallet.Chain,
		&i.Wallet.Symbol,
		&i.Wallet.Label,
		&i.Wallet.Notify,
		&i.Wallet.UserID,
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
		Address: in.Address,
		Chain:   in.Chain,
		Symbol:  in.Symbol,
	}
}

func userWalletToDomain(in models.Wallet) domain.UserWallet {
	return domain.UserWallet{
		UserID: in.UserID,
		Label:  in.Label,
		Notify: in.Notify,
		Wallet: domain.Wallet{
			ID:      in.ID.String(),
			Address: in.Address,
			Chain:   in.Chain,
			Symbol:  in.Symbol,
		},
	}
}

func walletToDomainBalance(in *models.WalletBalance) domain.UserWalletBalance {
	hasError := false
	var errorMsg string
	if !in.BalanceUSD.Valid || !in.Balance.Valid {
		hasError = true
		errorMsg = "Unable to fetch live balance"
	}
	return domain.UserWalletBalance{
		UserWallet: domain.UserWallet{
			UserID: in.Wallet.UserID,
			Label:  in.Wallet.Label,
			Notify: in.Wallet.Notify,
			Wallet: domain.Wallet{
				ID:      in.Wallet.ID.String(),
				Address: in.Wallet.Address,
				Chain:   in.Wallet.Chain,
				Symbol:  in.Wallet.Symbol,
			},
		},
		Balance:    in.Balance.Float64,
		BalanceUSD: in.BalanceUSD.Float64,
		UpdatedAt:  in.UpdatedAt.Time,
		HasError:   hasError,
		ErrorMsg:   errorMsg,
	}
}
