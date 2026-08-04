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

func (r *WalletRepository) ListWallets(ctx context.Context, userID string) ([]domain.Wallet, error) {
	query := `SELECT id, user_id, address, chain, symbol, label, created_at, updated_at FROM wallets ORDER BY created_at ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []domain.Wallet
	for rows.Next() {
		var wallet models.Wallet
		if err := rows.Scan(
			&wallet.ID,
			&wallet.UserID,
			&wallet.Address,
			&wallet.Chain,
			&wallet.Symbol,
			&wallet.Label,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, err
		}
		wallets = append(wallets, walletToDomain(wallet))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *WalletRepository) CreateWallet(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.Wallet, error) {
	query := `INSERT INTO wallets (address, chain, symbol, label, user_id)
        VALUES ($1, $2, $3, $4, $5)
		RETURNING *`
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
		&wallet.CreatedAt,
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

func (r *WalletRepository) EditWallet(ctx context.Context, userID string, id uuid.UUID, label string) (*domain.Wallet, error) {
	query := `UPDATE wallets
		SET label = $1
		WHERE user_id = $2 AND id = $3
		RETURNING*;`
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
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	); err != nil {
		return nil, core.ErrWalletNotFound
	}
	out := walletToDomain(wallet)
	return &out, nil
}

func (r *WalletRepository) DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error {
	query := `DELETE FROM wallets WHERE id = $1`
	res, err := r.db.Pool.Exec(ctx, query, id)
	if res.RowsAffected() == 0 {
		return core.ErrWalletNotFound
	}
	return err
}

func (r *WalletRepository) GetWallet(ctx context.Context, userID string, id uuid.UUID) (*domain.Wallet, error) {
	query := `SELECT * FROM wallets WHERE user_id = $1 AND id = $2`
	rows, err := r.db.Pool.Query(ctx, query, userID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var wallet models.Wallet
		if err := rows.Scan(
			&wallet.ID,
			&wallet.UserID,
			&wallet.Address,
			&wallet.Chain,
			&wallet.Symbol,
			&wallet.Label,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out := walletToDomain(wallet)
		return &out, nil
	}
	return nil, core.ErrWalletNotFound
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
