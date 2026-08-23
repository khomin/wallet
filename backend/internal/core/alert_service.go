package core

import (
	"context"
	"tracker/internal/core/domain"

	"github.com/sirupsen/logrus"
)

type AlertService struct {
	alertRepo      AlertRepository
	userRepo       UserRepo
	priceCache     PriceCache
	onNotification func(cmd domain.NotificationCommand) error
	log            *logrus.Entry
}

func NewAlertService(
	alertRepo AlertRepository,
	userRepo UserRepo,
	priceCache PriceCache,
	onNotification func(cmd domain.NotificationCommand) error,
) *AlertService {
	return &AlertService{
		alertRepo:      alertRepo,
		userRepo:       userRepo,
		priceCache:     priceCache,
		onNotification: onNotification,
		log:            logrus.WithField("component", "AlertService"),
	}
}

func (s *AlertService) ProcessAlerts(ctx context.Context) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		s.log.WithError(err).Error("failed to fetch users")
		return
	}
	for _, user := range users {
		s.processUserAlerts(ctx, user)
	}
}

func (s *AlertService) processUserAlerts(ctx context.Context, user domain.User) {
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

func (s *AlertService) isTriggered(alert domain.Alert, price *domain.TokenPrice) bool {
	switch alert.Condition {
	case domain.AlertConditionAbove:
		return price.GreaterThanOrEqual(alert.Price)
	case domain.AlertConditionBelow:
		return price.LessThanOrEqual(alert.Price)
	default:
		return false
	}
}

func (s *AlertService) triggerAlert(ctx context.Context, user domain.User, alert domain.Alert, price *domain.TokenPrice) {
	log := s.log.WithField("alert_id", alert.ID)
	log.Infof("Alert triggered for %s", alert.CoinSymbol)

	if s.onNotification != nil {
		coin := s.priceCache.GetCoinBySymbol(ctx, alert.CoinSymbol)
		coinName := "<name>"
		if coin != nil {
			coinName = coin.Name
		}
		err := s.onNotification(domain.NotificationCommand{
			UserID:     user.Name,
			CoinName:   coinName,
			CoinSymbol: alert.CoinSymbol,
			Email:      user.Email,
			UserName:   user.Name,
			AlertID:    alert.ID,
			Price:      price.CurrentPrice,
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
