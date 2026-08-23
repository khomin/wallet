package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
				subject := h.subject(cmd.CoinSymbol, cmd.Price)
				htmlBody, err := h.renderAlertTemplate(cmd.UserName, cmd.CoinSymbol, cmd.Price)
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

func (w *EmailWorker) renderAlertTemplate(userName, coinSymbol string, price float64) (string, error) {
	cleanName := strings.TrimSpace(userName)
	if cleanName == "" {
		cleanName = "there"
	}
	data := alertTemplateData{
		UserName:       cleanName,
		CoinSymbol:     strings.ToUpper(strings.TrimSpace(coinSymbol)),
		FormattedPrice: fmt.Sprintf("%.2f", price),
	}
	var buf bytes.Buffer
	if err := alertEmailTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}
	return buf.String(), nil
}

func (h *EmailWorker) subject(name string, price float64) string {
	return fmt.Sprintf("🚨 %s Price Alert Triggered (%v)", name, price)
}

type alertTemplateData struct {
	UserName       string
	CoinSymbol     string
	FormattedPrice string
}
