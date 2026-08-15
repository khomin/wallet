package core

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"
	"tracker/internal/core/domain"
	"tracker/internal/messaging"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type walletWorker struct {
	eventConsumer *messaging.Consumer
	walletService *WalletService
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
	}
}

func (w *walletWorker) StartConsuming(ctx context.Context) error {
	log := logrus.WithField("walletWorker", "StartConsuming")

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
			s.SynchronizeWallets(ctx)
		}
	}
}

func (w *walletWorker) handleWalletCreated(ctx context.Context, msg amqp091.Delivery) error {
	var event domain.WalletCreatedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return nil
	}
	walletID, err := uuid.Parse(event.ID)
	if err != nil {
		return nil
	}
	wallet, err := w.walletRepo.Get(ctx, event.UserID, walletID)
	if err != nil {
		return err
	}
	err = w.synchronizeWallet(ctx, wallet.Wallet)
	if err != nil {
		return err
	}
	return nil
}

func (s *walletWorker) synchronizeWallet(ctx context.Context, wallet domain.Wallet) error {
	// FIXME: don't think i should keep it here
	balance, err := s.walletService.FetchBalance(ctx, wallet)
	if err != nil {
		return err
	}
	uuid, err := uuid.Parse(wallet.ID)
	if err != nil {
		return err
	}
	err = s.walletRepo.UpdateBalance(ctx, wallet.UserID, uuid, balance.Balance, balance.BalanceUSD)
	if err != nil {
		return err
	}
	return nil
}

func (s *walletWorker) SynchronizeWallets(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	wallets, err := s.walletRepo.ListForSync(ctx, 100)
	if err != nil {
		slog.Warn("failed to fetch active wallets for sync", "error", err)
		return err
	}
	groupsByChain := map[string][]domain.Wallet{}
	for _, wallet := range wallets {
		groupsByChain[wallet.Chain] = append(groupsByChain[wallet.Chain], wallet)
	}
	for _, group := range groupsByChain {
		group := group
		g.Go(func() error {
			for _, wallet := range group {
				if err := s.handleOneBalance(ctx, wallet); err != nil {
					slog.Info("WALLET_ERROR", "chain", wallet.Chain, "wallet", wallet.ID, "error", err)
					continue
				}
				slog.Info("WALLET_UPDATED", "chain", wallet.Chain, "wallet", wallet.ID)
			}
			return nil
		})
	}
	err = g.Wait()
	slog.Info("WALLET_SYNC_END", "len", len(wallets))
	return err
}

func (s *walletWorker) handleOneBalance(ctx context.Context, wallet domain.Wallet) error {
	uuid, err := uuid.Parse(wallet.ID)
	if err != nil {
		return err
	}
	balance, err := s.walletService.FetchBalance(ctx, wallet)
	if err != nil {
		if errors.Is(err, ErrProviderTimeout) {
			return err
		}
		if errors.Is(err, ErrProviderRateLimit) {
			return err
		}
		if errors.Is(err, ErrProviderUnavailable) {
			return err
		}
		return err
	}
	err = s.walletRepo.UpdateBalance(ctx, wallet.UserID, uuid, balance.Balance, balance.BalanceUSD)
	if err != nil {
		return err
	}
	return nil
}
