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
	ch            *amqp091.Channel
	walletService *WalletService
}

func NewWalletWorker(walletService *WalletService) *WalletWorker {
	return &WalletWorker{
		walletService: walletService,
		ch:            &amqp091.Channel{},
	}
}

func (w *WalletWorker) StartConsuming(ctx context.Context) error {
	msgs, err := w.ch.Consume(messaging.QueueWalletCreated, "wallet-worker-1", false, false, false, false, nil)
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
		return err
	}
	w.walletService.FetchPortfolio(ctx, domain.Wallet{
		ID:     event.ID,
		UserID: event.UserID,
	})
	// print(event)
	return nil
}
