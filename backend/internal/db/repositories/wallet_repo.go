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
	"github.com/jackc/pgx/v5/pgconn"
)

type walletRepository struct {
	db *db.DataBase
}

func NewWalletRepository(db *db.DataBase) core.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) List(ctx context.Context, userID string) ([]domain.WalletBalance, error) {
	query := `SELECT
		w.id, w.user_id,  w.address, w.chain, coin.symbol, w.label, w.updated_at,
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
	ORDER BY w.updated_at ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []domain.WalletBalance
	for rows.Next() {
		var w models.WalletBalance
		if err := rows.Scan(
			&w.Wallet.ID,
			&w.Wallet.UserID,
			&w.Wallet.Address,
			&w.Wallet.Chain,
			&w.Wallet.Symbol,
			&w.Wallet.Label,
			&w.Wallet.UpdatedAt,
			//
			&w.Balance,
			&w.BalanceUSD,
			&w.BalanceUpdatedAt,
			//
			&w.Price.ID,
			&w.Price.Symbol,
			&w.Price.Name,
			&w.Price.CurrentPrice,
			&w.Price.MarketCap,
			&w.Price.TotalVolume,
			&w.Price.Change_24h,
			&w.Price.PriceChangePercentage_24h,
			&w.Price.MarketCapChange_24h,
			&w.Price.MarketCapChange_percentage_24h,
			&w.Price.UpdatedAt,
		); err != nil {
			return nil, err
		}
		wallets = append(wallets, walletToDomain2(w))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *walletRepository) Get(ctx context.Context, userID string, id uuid.UUID) (*domain.WalletBalance, error) {
	query := `SELECT
			w.id, w.user_id,  w.address, w.chain, coin.symbol, w.label, w.updated_at,
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
		WHERE w.user_id = $1 AND w.id = $2`
	rows, err := r.db.Pool.Query(ctx, query,
		userID,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var w models.WalletBalance
		if err := rows.Scan(
			&w.Wallet.ID,
			&w.Wallet.UserID,
			&w.Wallet.Address,
			&w.Wallet.Chain,
			&w.Wallet.Symbol,
			&w.Wallet.Label,
			&w.Wallet.UpdatedAt,
			//
			&w.Balance,
			&w.BalanceUSD,
			&w.BalanceUpdatedAt,
			//
			&w.Price.ID,
			&w.Price.Symbol,
			&w.Price.Name,
			&w.Price.CurrentPrice,
			&w.Price.MarketCap,
			&w.Price.TotalVolume,
			&w.Price.Change_24h,
			&w.Price.PriceChangePercentage_24h,
			&w.Price.MarketCapChange_24h,
			&w.Price.MarketCapChange_percentage_24h,
			&w.Price.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out := walletToDomain2(w)
		return &out, nil
	}
	return nil, domain.ErrorNotFound
}

func (r *walletRepository) Create(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.Wallet, error) {
	query := `INSERT INTO wallets (address, chain, coin_id, label, user_id)
		VALUES (
			$1, 
			$2,
			(SELECT id FROM coins WHERE symbol = $3), 
			$4, $5
		)
		RETURNING id, user_id, address, chain, (SELECT symbol FROM coins WHERE symbol = $3), label, updated_at ;`
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

func (r *walletRepository) Edit(ctx context.Context, userID string, id uuid.UUID, label string) (*domain.Wallet, error) {
	query := `UPDATE wallets SET label = $1
		WHERE user_id = $2 AND id = $3
		RETURNING id, user_id, address, chain, symbol, label, updated_at;`
	row := r.db.Pool.QueryRow(ctx, query,
		label,
		userID, id,
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

func (r *walletRepository) UpdateBalance(ctx context.Context, userID string, id uuid.UUID, crypto float64, usd float64) error {
	query := `INSERT INTO wallet_balances 
		(id, value_crypto, value_usd, updated_at)
		VALUES ($1, $2, $3, NOW())
	ON CONFLICT (id)
	DO UPDATE SET value_crypto = EXCLUDED.value_crypto, value_usd = EXCLUDED.value_usd, updated_at = EXCLUDED.updated_at
	RETURNING id, value_crypto, value_usd, updated_at`

	_, err := r.db.Pool.Exec(ctx, query,
		id, crypto, usd,
	)
	if err != nil {
		return err
	}
	return err
}

func (r *walletRepository) ListForSync(ctx context.Context, limit int) ([]domain.Wallet, error) {
	query := `SELECT 
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
	LIMIT $2`

	rows, err := r.db.Pool.Query(ctx,
		query,
		time.Now().Add(-5*time.Minute),
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

func walletToDomain(in models.Wallet) domain.Wallet {
	return domain.Wallet{
		ID:      in.ID.String(),
		Address: in.Address,
		Chain:   in.Chain,
		Label:   in.Label,
		Symbol:  in.Symbol,
		UserID:  in.UserID,
	}
}

func walletToDomain2(in models.WalletBalance) domain.WalletBalance {
	hasError := false
	var errorMsg string
	if !in.BalanceUSD.Valid || !in.Balance.Valid {
		hasError = true
		errorMsg = "Unable to fetch live balance"
	}
	return domain.WalletBalance{
		Wallet: domain.Wallet{
			ID:      in.Wallet.ID.String(),
			Address: in.Wallet.Address,
			Chain:   in.Wallet.Chain,
			Label:   in.Wallet.Label,
			Symbol:  in.Wallet.Symbol,
			UserID:  in.Wallet.UserID,
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
		HasError:   hasError,
		ErrorMsg:   errorMsg,
	}
}
