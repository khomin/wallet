package demo

import (
	"time"
	"tracker/internal/core/domain"

	"github.com/google/uuid"
)

type DemoAlerts struct {
	Alerts map[string]domain.Alert
}

func NewDemoAlerts() *DemoAlerts {
	var alerts = make(map[string]domain.Alert)
	for _, i := range alertList {
		alerts[i.ID] = i
	}
	return &DemoAlerts{
		Alerts: alerts,
	}
}

func (d *DemoAlerts) GetAlerts() []domain.Alert {
	return alertList
}

func (d *DemoAlerts) GetAlert(id uuid.UUID) (*domain.Alert, error) {
	v, found := d.Alerts[id.String()]
	if !found {
		return nil, domain.ErrorNotFound
	}
	return &v, nil
}

var alertList = []domain.Alert{
	{
		ID:          "9248FE7A-7258-4371-B24D-FC4C11DEDB00",
		UserID:      "demo",
		CoinSymbol:  "BTC",
		Condition:   domain.AlertConditionAbove,
		Price:       85000,
		Enabled:     true,
		TriggeredAt: nil,
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now().Add(-time.Hour * 24),
	},
	{
		ID:          "9248FE7A-7258-4371-B24D-FC4C11DEDB01",
		UserID:      "demo",
		CoinSymbol:  "TRX",
		Condition:   domain.AlertConditionAbove,
		Price:       34.0,
		Enabled:     true,
		TriggeredAt: nil,
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now().Add(-time.Hour * 24),
	},
	{
		ID:          "9248FE7A-7258-4371-B24D-FC4C11DEDB02",
		UserID:      "demo",
		CoinSymbol:  "SOL",
		Condition:   domain.AlertConditionAbove,
		Price:       100,
		Enabled:     true,
		TriggeredAt: nil,
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now().Add(-time.Hour * 24),
	},
}
