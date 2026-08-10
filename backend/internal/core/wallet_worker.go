package core

import (
	"context"
	"encoding/json"
	"time"
	"tracker/internal/core/domain"
	"tracker/internal/messaging"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type walletWorker struct {
	eventConsumer *messaging.Consumer
	walletService *WalletService
	rateLimiter   *rate.Limiter
	walletRepo    WalletRepository
}

type NewWalletDeps struct {
	WalletService *WalletService
	WalletRepo    WalletRepository
	MqConsumer    *messaging.Consumer
}

func NewWalletWorker(deps *NewWalletDeps) *walletWorker {
	return &walletWorker{
		walletRepo:    deps.WalletRepo,
		eventConsumer: deps.MqConsumer,
		walletService: deps.WalletService,
		rateLimiter:   rate.NewLimiter(rate.Limit(10), 1),
	}
}

func (w *walletWorker) StartConsuming(ctx context.Context) error {
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

func (s *walletWorker) StartSyncLoop(ctx context.Context, interval time.Duration) {
	tm := time.NewTicker(interval)
	defer tm.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			s.walletService.SynchronizeWallets(ctx)
		}
	}
}

func (w *walletWorker) handleWalletCreated(ctx context.Context, msg amqp091.Delivery) error {
	var wallet domain.WalletCreatedEvent
	if err := json.Unmarshal(msg.Body, &wallet); err != nil {
		return nil
	}
	err := w.walletService.SynchronizeWallet(ctx, domain.Wallet{
		ID:     wallet.ID,
		UserID: wallet.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}
