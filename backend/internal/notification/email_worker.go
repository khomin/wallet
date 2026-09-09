package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"tracker/internal/core/domain"
	"tracker/internal/messaging"

	"github.com/sirupsen/logrus"
)

type EmailWorker struct {
	consumer *messaging.Consumer
	sender   EmailSender
}

func NewEmailWorker(consumer *messaging.Consumer, sender EmailSender) *EmailWorker {
	return &EmailWorker{
		consumer: consumer,
		sender:   sender,
	}
}

func (h *EmailWorker) Start(ctx context.Context) {
	log := logrus.WithField("EmailWorker", "Start")

	for {
		deliveries, closeChan, err := h.consumer.Consume()
		if err != nil {
			log.Warnf("consume error: %v, retrying...", err)
			continue
		}
	loop:
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-closeChan:
				log.Debugf("Channel closed: %v. Reconnecting...", err)
				break loop
			case d, ok := <-deliveries:
				if !ok {
					break
				}
				var cmd domain.NotificationCommand
				if err := json.Unmarshal(d.Body, &cmd); err != nil {
					log.Debugf("Failed to unmarshal event: %v", err)
					_ = d.Nack(false, false)
					continue
				}
				switch cmd.Type {
				case domain.NotificationPriceAlert:
					var event domain.PriceAlertNotification
					err := json.Unmarshal(cmd.Payload, &event)
					if err != nil {
						log.WithError(err).Error("failed to unmarshal event")
						return
					}
					subject := SubjectPrice(event.CoinSymbol, event.Price)
					htmlBody, err := RenderAlertTemplate(cmd.UserName, event.CoinSymbol, event.Price)
					if err != nil {
						log.WithError(err).Error("failed format template")
						return
					}
					err = h.sender.Send(ctx, cmd.Email, subject, htmlBody)
					if err != nil {
						log.WithError(err).Error("failed to send email")
						continue
					} else {
						_ = d.Ack(false)
					}
				case domain.NotificationBalanceAlert:
					var event domain.BalanceAlertNotification
					err := json.Unmarshal(cmd.Payload, &event)
					if err != nil {
						log.WithError(err).Error("failed to unmarshal event")
						return
					}
					subject := SubjectBalance(event.Name, event.Current)
					formattedDelta := fmt.Sprintf("%+.4f", event.Delta)
					formattedBalance := fmt.Sprintf("%.4f", event.Current)

					htmlBody, err := RenderBalanceEmail(BalanceTemplateData{
						UserName:         cmd.UserName,
						WalletName:       event.Name,
						CoinSymbol:       event.CoinSymbol,
						FormattedBalance: formattedBalance,
						FormattedDelta:   formattedDelta,
						IsDeposit:        event.Delta > 0,
					})
					if err != nil {
						log.WithError(err).Error("failed format template")
						return
					}
					err = h.sender.Send(ctx, cmd.Email, subject, htmlBody)
					if err != nil {
						log.WithError(err).Error("failed to send email")
						continue
					} else {
						_ = d.Ack(false)
					}
				}
			}
		}
	}
}

type alertTemplateData struct {
	UserName       string
	CoinSymbol     string
	FormattedPrice string
}

type balanceTemplateData struct {
	UserName         string
	WalletName       string
	FormattedBalance string
}
