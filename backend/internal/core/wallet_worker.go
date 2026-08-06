package core

import (
	"context"
	"encoding/json"
	"tracker/internal/core/domain"
	"tracker/internal/messaging"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type WalletWorker struct {
	mqConsumer    *messaging.Consumer
	walletService *WalletService
}

func NewWalletWorker(walletService *WalletService, mqConsumer *messaging.Consumer) *WalletWorker {
	return &WalletWorker{
		walletService: walletService,
		mqConsumer:    mqConsumer,
	}
}

func (w *WalletWorker) StartConsuming(ctx context.Context) error {
	msgs, err := w.mqConsumer.Consume()
	if err != nil {
		return err
	}
	go func() {
		for msg := range msgs {
			if err := w.handleWalletCreated(ctx, msg); err != nil {
				logrus.WithError(err).Error("failed to handle wallet created event")
				_ = msg.Nack(false, true)
			} else {
				_ = msg.Ack(false)
			}
		}
	}()
	return nil
}

func (w *WalletWorker) handleWalletCreated(ctx context.Context, msg amqp091.Delivery) error {
	var event domain.WalletCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return nil
	}
	w.walletService.FetchPortfolio(ctx, domain.Wallet{
		ID:     event.ID,
		UserID: event.UserID,
	})
	return nil
}
