package repositories

import (
	"context"
	"errors"
	"strings"
	"tracker/internal/core"
	"tracker/internal/core/domain"
	"tracker/internal/db"
	"tracker/internal/db/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type WalletRepository struct {
	db *db.DataBase
}

func NewWalletRepository(db *db.DataBase) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) List(ctx context.Context, userID string) ([]domain.WalletWithBalance, error) {
	query := `SELECT
		w.id, w.user_id,  w.address, w.chain, coin.symbol, w.label, w.updated_at,
		balance.price,
		balance.price_usd,
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

	var wallets []domain.WalletWithBalance
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

func (r *WalletRepository) Get(ctx context.Context, userID string, id uuid.UUID) (*domain.WalletWithBalance, error) {
	query := `SELECT w.id, w.user_id, w.address, w.chain, w.label, w.updated_at,
		p.price_usd,
		p.market_cap_usd,
		p.total_volume_usd,
		p.price_change_24h,
		p.price_change_percent_24h,
		p.market_cap_change_24h,
		p.market_cap_change_percent_24h,
		p.updated_at		
		FROM wallets w
		JOIN coin_prices p ON p.id = w.coin_id
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
			&w.Balance,
			// &wallet.Price.Change_24h,//
			&w.Balance,
			&w.BalanceUSD,
			&w.HasError,
			&w.ErrorMsg,
		); err != nil {
			return nil, err
		}
		out := walletToDomain2(w)
		return &out, nil
	}
	return nil, core.ErrWalletNotFound
}

func (r *WalletRepository) Create(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.Wallet, error) {
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
			return nil, core.ErrWalletAlreadyExists
		}
		return nil, core.ErrWalletInternalError
	}
	out := walletToDomain(wallet)
	return &out, nil
}

func (r *WalletRepository) Edit(ctx context.Context, userID string, id uuid.UUID, label string) (*domain.Wallet, error) {
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
		return nil, core.ErrWalletNotFound
	}
	out := walletToDomain(wallet)
	return &out, nil
}

func (r *WalletRepository) Delete(ctx context.Context, userID string, id uuid.UUID) error {
	query := `DELETE FROM wallets WHERE id = $1`
	res, err := r.db.Pool.Exec(ctx, query, id)
	if res.RowsAffected() == 0 {
		return core.ErrWalletNotFound
	}
	return err
}

func (r *WalletRepository) SetBalance(ctx context.Context, userID string, id uuid.UUID, balance float64, balanceUSD float64) error {
	query := `INSERT INTO wallet_balances (id, price, price_usd, updated_at)
		VALUES ($1, 1230, 45611, NOW())
		ON CONFLICT (id)
		DO UPDATE SET price = EXCLUDED.price, price_usd = EXCLUDED.price_usd, updated_at = EXCLUDED.updated_at
		RETURNING id, price, price_usd, updated_at`

	res, err := r.db.Pool.Exec(ctx, query, id)
	if res.RowsAffected() == 0 {
		return core.ErrWalletNotFound
	}
	return err
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

func (r *WalletRepository) ListForSync(ctx context.Context, limit int) ([]domain.Wallet, error) {
	// query := `SELECT id FROM wallet_balances (id, price, price_usd, updated_at)
	// 	VALUES ($1, 1230, 45611, NOW())
	// 	ON CONFLICT (id)
	// 	DO UPDATE SET price = EXCLUDED.price, price_usd = EXCLUDED.price_usd, updated_at = EXCLUDED.updated_at
	// 	RETURNING id, price, price_usd, updated_at`

	// res, err := r.db.Pool.Exec(ctx, query, id)
	// if res.RowsAffected() == 0 {
	// 	return core.ErrWalletNotFound
	// }
	// return err
	return nil, nil
}

func walletToDomain2(in models.WalletBalance) domain.WalletWithBalance {
	hasError := false
	var errorMsg string
	if !in.BalanceUSD.Valid || !in.Balance.Valid {
		hasError = true
		errorMsg = "Unable to fetch live balance"
	}
	return domain.WalletWithBalance{
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
