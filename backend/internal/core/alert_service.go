package core

import (
	"context"
	"fmt"
	"tracker/internal/core/domain"

	"github.com/sirupsen/logrus"
)

type EmailSender interface {
	Send(ctx context.Context, recipient, subject, body string) error
}

type AlertService struct {
	alertRepo   AlertRepository
	userRepo    UserRepo
	priceCache  PriceCache
	emailSender EmailSender
	log         *logrus.Entry
}

func NewAlertService(
	alertRepo AlertRepository,
	userRepo UserRepo,
	priceCache PriceCache,
	emailSender EmailSender,
) *AlertService {
	return &AlertService{
		alertRepo:   alertRepo,
		userRepo:    userRepo,
		priceCache:  priceCache,
		emailSender: emailSender,
		log:         logrus.WithField("component", "AlertService"),
	}
}

func (s *AlertService) EvaluateAlerts(ctx context.Context) {
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

func (s *AlertService) triggerAlert(ctx context.Context, user domain.User, alert domain.Alert, _ *domain.TokenPrice) {
	s.log.WithField("alert_id", alert.ID).Infof("Alert triggered for %s", alert.CoinSymbol)

	if s.emailSender != nil {
		subject := fmt.Sprintf("%s price alert triggered", alert.CoinSymbol)
		body := fmt.Sprintf("Hello %s,\n\nYour %s price alert was triggered.", user.Name, alert.CoinSymbol)

		// TODO: push this to a RabbitMQ queue instead of blocking
		if err := s.emailSender.Send(ctx, user.Email, subject, body); err != nil {
			s.log.WithError(err).Error("failed to send alert email")
			return
		}
	}
	if _, err := s.alertRepo.Disable(ctx, user.ID, alert.ID); err != nil {
		s.log.WithError(err).Error("failed to disable triggered alert")
	}
}
