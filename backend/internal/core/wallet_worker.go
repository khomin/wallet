package core

import (
	"context"
	"encoding/json"
	"time"
	"tracker/internal/core/domain"
	"tracker/internal/messaging"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type WalletWorker struct {
	eventConsumer *messaging.Consumer
	walletService *WalletService
}

func NewWalletWorker(walletService *WalletService, mqConsumer *messaging.Consumer) *WalletWorker {
	return &WalletWorker{
		walletService: walletService,
		eventConsumer: mqConsumer,
	}
}

func (w *WalletWorker) StartConsuming(ctx context.Context) error {
	log := logrus.WithField("walletWorker", "StartConsuming")
	go func() {
		for {
			msgs, chanErr, err := w.eventConsumer.Consume()
			if err != nil {
				log.Warningf("failed to set consume: %v", err)
				time.Sleep(time.Second * 1)
				continue
			}
			log.Info("event consume set successfully")
			for {
				var exitErr error
				select {
				case err := <-chanErr:
					log.Infof("channel closed: %v", err)
					exitErr = err
				case msg := <-msgs:
					log.Infof("event received")
					if err := w.handleWalletCreated(ctx, msg); err != nil {
						logrus.WithError(err).Error("failed to handle wallet created event")
						_ = msg.Nack(false, true)
					} else {
						_ = msg.Ack(false)
					}
				}
				if exitErr != nil {
					log.Infof("channel closed due to error: %v", err)
					break
				}
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
