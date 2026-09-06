package core

import (
	"context"
	"encoding/json"
	"tracker/internal/core/domain"

	"github.com/sirupsen/logrus"
)

type NotificationService struct {
	alertRepo      AlertRepository
	userRepo       UserRepo
	priceCache     PriceCache
	onNotification func(cmd domain.NotificationCommand) error
	log            *logrus.Entry
}

func NewNotificationService(
	alertRepo AlertRepository,
	userRepo UserRepo,
	priceCache PriceCache,
	onNotification func(cmd domain.NotificationCommand) error,
) *NotificationService {
	return &NotificationService{
		alertRepo:      alertRepo,
		userRepo:       userRepo,
		priceCache:     priceCache,
		onNotification: onNotification,
		log:            logrus.WithField("component", "AlertService"),
	}
}

func (s *NotificationService) ProcessAlerts(ctx context.Context) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		s.log.WithError(err).Error("failed to fetch users")
		return
	}
	for _, user := range users {
		s.processUserAlerts(ctx, user)
	}
}

func (s *NotificationService) WalletBalanceChanged(ctx context.Context, balance domain.WalletBalanceChange) {
	user, err := s.userRepo.GetByID(ctx, balance.Wallet.UserID)
	if err != nil {
		s.log.WithError(err).Error("failed to fetch user")
		return
	}
	s.triggerBalanceAlert(ctx, user, balance)
}

func (s *NotificationService) processUserAlerts(ctx context.Context, user domain.User) {
	alerts, err := s.alertRepo.ListByUser(ctx, user.ID)
	if err != nil {
		s.log.WithError(err).Error("failed to fetch alerts")
		return
	}
	for _, alert := range alerts {
		if !alert.Enabled {
			continue
		}
		price := s.priceCache.GetPriceBySymbol(ctx, alert.CoinSymbol)
		if price == nil {
			continue
		}
		if s.isTriggered(alert, price) {
			s.triggerAlert(ctx, user, alert, price)
		}
	}
}

func (s *NotificationService) isTriggered(alert domain.Alert, price *domain.TokenPrice) bool {
	switch alert.Condition {
	case domain.AlertConditionAbove:
		return price.GreaterThanOrEqual(alert.Price)
	case domain.AlertConditionBelow:
		return price.LessThanOrEqual(alert.Price)
	default:
		return false
	}
}

func (s *NotificationService) triggerAlert(ctx context.Context, user domain.User, alert domain.Alert, price *domain.TokenPrice) {
	log := s.log.WithField("alert_id", alert.ID)
	log.Infof("Alert triggered for %s", alert.CoinSymbol)

	if s.onNotification != nil {
		coin := s.priceCache.GetCoinBySymbol(ctx, alert.CoinSymbol)
		coinName := "<name>"
		if coin != nil {
			coinName = coin.Name
		}
		alert := domain.PriceAlertNotification{
			CoinName:   coinName,
			CoinSymbol: alert.CoinSymbol,
			AlertID:    alert.ID,
			Price:      price.CurrentPrice,
		}
		raw, err := json.Marshal(alert)
		if err != nil {
			log.WithError(err).Error("failed marshal notification")
			return
		}
		err = s.onNotification(domain.NotificationCommand{
			UserID:   user.Name,
			Type:     domain.NotificationPriceAlert,
			Payload:  raw,
			Email:    user.Email,
			UserName: user.Name,
		})
		if err != nil {
			log.WithError(err).Error("failed to send alert email")
			return
		}
	}
	if _, err := s.alertRepo.DisableAsCompleted(ctx, user.ID, alert.ID); err != nil {
		log.WithError(err).Error("failed to disable triggered alert")
	}
}

func (s *NotificationService) triggerBalanceAlert(_ context.Context, user *domain.User, balance domain.WalletBalanceChange) {
	log := s.log.WithField("balance_id", balance.ID)
	log.Infof("Balance alert triggered for %s", balance.Address)

	if s.onNotification != nil {
		alert := domain.BalanceAlertNotification{
			WalletName: balance.Label,
			CoinSymbol: balance.Symbol,
			Current:    balance.CurrentBalance,
			Previous:   balance.OldBalance,
			Delta:      balance.CurrentBalance - balance.OldBalance,
		}
		raw, err := json.Marshal(alert)
		if err != nil {
			log.WithError(err).Error("failed marshal notification")
			return
		}
		err = s.onNotification(domain.NotificationCommand{
			UserID:   user.Name,
			Type:     domain.NotificationBalanceAlert,
			Payload:  raw,
			Email:    user.Email,
			UserName: user.Name,
		})
		if err != nil {
			log.WithError(err).Error("failed to send alert email")
			return
		}
	}
}
