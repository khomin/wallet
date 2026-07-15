package repositories

import (
	"context"
	"tracker/internal/core"
	"tracker/internal/db"
	"tracker/internal/db/models"

	"github.com/google/uuid"
)

type WalletRepository struct {
	db *db.DataBase
}

func NewWalletRepository(db *db.DataBase) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) ListWallets(ctx context.Context, userID string) ([]models.Wallet, error) {
	query := `SELECT id, address, chain, label, user_id, created_at, updated_at FROM wallets ORDER BY created_at ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []models.Wallet
	for rows.Next() {
		var wallet models.Wallet
		if err := rows.Scan(
			&wallet.ID,
			&wallet.Address,
			&wallet.Chain,
			&wallet.Label,
			&wallet.UserID,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *WalletRepository) CreateWallet(ctx context.Context, userID string, chain string, address string, label string) (*models.Wallet, error) {
	query := `INSERT INTO wallets (address, chain, label, user_id)
        VALUES ($1, $2, $3, $4)
		RETURNING *`
	rows, err := r.db.Pool.Query(ctx, query,
		address,
		chain,
		label,
		userID,
	)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var wallet models.Wallet
		if err := rows.Scan(
			&wallet.ID,
			&wallet.Address,
			&wallet.Chain,
			&wallet.Label,
			&wallet.UserID,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, err
		}

	}
	return nil, core.ErrDatabase
}

func (r *WalletRepository) DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error {
	query := `DELETE FROM wallets WHERE id = $1`
	res, err := r.db.Pool.Exec(ctx, query, id)
	if res.RowsAffected() == 0 {
		return core.ErrWalletNotFound
	}
	return err
}

func (r *WalletRepository) GetWallet(ctx context.Context, userID string, id uuid.UUID) (*models.Wallet, error) {
	query := `SELECT * FROM wallets WHERE id = $1 AND user_id = $2`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var wallet models.Wallet
		if err := rows.Scan(
			&wallet.ID,
			&wallet.Address,
			&wallet.Chain,
			&wallet.Label,
			&wallet.UserID,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, err
		}
		return &wallet, nil
	}
	return nil, core.ErrWalletNotFound
}
