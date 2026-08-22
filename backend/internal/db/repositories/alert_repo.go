package repositories

import (
	"context"
	"strings"
	"tracker/internal/core"
	"tracker/internal/core/domain"
	"tracker/internal/db"
	"tracker/internal/db/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AlertRepository struct {
	db *db.DataBase
}

func NewAlertRepository(db *db.DataBase) core.AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) ListByUser(ctx context.Context, userID string) ([]domain.Alert, error) {
	query := `SELECT
		id, user_id,
			(SELECT symbol FROM coins WHERE id = coin_id), 
		condition, price, enabled, triggered_at, created_at, updated_at 
	FROM alerts 
	WHERE user_id = $1
	ORDER BY created_at ASC`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := make([]domain.Alert, 0)
	for rows.Next() {
		out, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alertToDomain(out))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return alerts, nil
}

func (r *AlertRepository) ListActive(ctx context.Context) ([]domain.Alert, error) {
	query := `SELECT 
		id, user_id,
			(SELECT symbol FROM coins WHERE id = coin_id),
		condition, price, enabled, triggered_at, created_at, updated_at
	FROM alerts
	WHERE enabled = TRUE
	ORDER BY created_at ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := make([]domain.Alert, 0)
	for rows.Next() {
		out, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alertToDomain(out))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return alerts, nil
}

func (r *AlertRepository) Create(ctx context.Context, alert domain.Alert) (*domain.Alert, error) {
	query := `
		INSERT INTO alerts (user_id, coin_id, condition, price)
		VALUES ($1, (SELECT id FROM coins WHERE symbol = $2), $3, $4)
		RETURNING 
			id, user_id,
				(SELECT symbol FROM coins WHERE id = coin_id), 
			condition, price, enabled, triggered_at, created_at, updated_at
		`
	row := r.db.Pool.QueryRow(ctx, query,
		alert.UserID,
		strings.ToUpper(alert.CoinSymbol),
		strings.ToLower(alert.Condition),
		alert.Price,
	)
	out, err := scanAlert(row)
	if err != nil {
		return nil, err
	}
	created := alertToDomain(out)
	return &created, nil
}

func (r *AlertRepository) Update(ctx context.Context, userID, id string, alert domain.AlertUpdate) (*domain.Alert, error) {
	query := `
		UPDATE alerts 
		SET 
			condition = $3,
			price = $4
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, (SELECT symbol FROM coins WHERE id = coin_id), condition, price, enabled, triggered_at, created_at, updated_at
	`
	row := r.db.Pool.QueryRow(ctx, query,
		userID,
		id,
		strings.ToLower(alert.Condition),
		alert.Price,
	)
	out, err := scanAlert(row)
	if err != nil {
		return nil, err
	}
	created := alertToDomain(out)
	return &created, nil
}

func (r *AlertRepository) Delete(ctx context.Context, userID string, id string) error {
	alertID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrInvalidArgument
	}
	res, err := r.db.Pool.Exec(ctx, `DELETE FROM alerts WHERE user_id = $1 AND id = $2`, userID, alertID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrAlertNotFound
	}
	return nil
}

func (r *AlertRepository) DisableAsCompleted(ctx context.Context, userID string, id string) (*domain.Alert, error) {
	alertID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}
	query := `
		UPDATE alerts
		SET
			enabled = FALSE,
			triggered_at = NOW(),
			updated_at = NOW()
		WHERE user_id = $1 AND id = $2 
		RETURNING id, user_id, (SELECT symbol FROM coins WHERE id = coin_id), condition, price, enabled, triggered_at, created_at, updated_at
	`
	row := r.db.Pool.QueryRow(ctx, query,
		userID,
		alertID,
	)
	out, err := scanAlert(row)
	if err != nil {
		return nil, err
	}
	created := alertToDomain(out)
	return &created, nil
}

func (r *AlertRepository) Pause(ctx context.Context, userID string, id string) (*domain.Alert, error) {
	alertID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}
	query := `
		UPDATE alerts
		SET 
			enabled = FALSE,
			updated_at = NOW(),
			triggered_at = NULL
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, (SELECT symbol FROM coins WHERE id = coin_id), condition, price, enabled, triggered_at, created_at, updated_at
	`
	row := r.db.Pool.QueryRow(ctx, query,
		userID,
		alertID,
	)
	out, err := scanAlert(row)
	if err != nil {
		return nil, err
	}
	created := alertToDomain(out)
	return &created, nil
}

func (r *AlertRepository) Enable(ctx context.Context, userID string, id string) (*domain.Alert, error) {
	alertID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}
	query := `
		UPDATE alerts
		SET 
			enabled = TRUE,
			updated_at = NOW(),
			triggered_at = NULL
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, (SELECT symbol FROM coins WHERE id = coin_id), condition, price, enabled, triggered_at, created_at, updated_at
	`
	row := r.db.Pool.QueryRow(ctx, query,
		userID,
		alertID,
	)
	out, err := scanAlert(row)
	if err != nil {
		return nil, err
	}
	created := alertToDomain(out)
	return &created, nil
}

func scanAlert(row pgx.Row) (models.Alert, error) {
	var alert models.Alert
	err := row.Scan(
		&alert.ID,
		&alert.UserID,
		&alert.CoinID,
		&alert.Condition,
		&alert.Price,
		&alert.Enabled,
		&alert.TriggeredAt,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)
	return alert, err
}

func alertToDomain(in models.Alert) domain.Alert {
	out := domain.Alert{
		ID:         in.ID.String(),
		UserID:     in.UserID,
		CoinSymbol: in.CoinID,
		Condition:  in.Condition,
		Price:      in.Price.Float64,
		Enabled:    in.Enabled,
		CreatedAt:  in.CreatedAt,
		UpdatedAt:  in.UpdatedAt,
	}
	if in.TriggeredAt.Valid {
		triggeredAt := in.TriggeredAt.Time
		out.TriggeredAt = &triggeredAt
	}
	return out
}
