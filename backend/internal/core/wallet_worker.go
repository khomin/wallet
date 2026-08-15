package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	start := time.Now()
	g, ctx := errgroup.WithContext(ctx)
	log := logrus.WithContext(ctx)
	groupsByChain := make(map[string][]domain.Wallet)

	wallets, err := s.walletRepo.ListForSync(ctx, 100)
	if err != nil {
		log.WithError(err).WithField("limit", 100).Error("failed to fetch wallets for sync")
		return fmt.Errorf("list wallets for sync: %w", err)
	}
	if len(wallets) == 0 {
		log.Debug("no wallets found for synchronization")
		return nil
	}

	for _, wallet := range wallets {
		groupsByChain[wallet.Chain] = append(groupsByChain[wallet.Chain], wallet)
	}

	log.WithFields(logrus.Fields{
		"total_wallets": len(wallets),
		"chain_count":   len(groupsByChain),
	}).Info("starting wallet synchronization batch")

	for chain, group := range groupsByChain {
		chain := chain
		group := group

		g.Go(func() error {
			var failedCount int

			for _, wallet := range group {
				if err := s.handleOneBalance(ctx, wallet); err != nil {
					failedCount++
					log.WithFields(logrus.Fields{
						"chain":     wallet.Chain,
						"wallet_id": wallet.ID,
					}).WithError(err).Error("failed to sync wallet balance")
					continue
				}
				log.WithFields(logrus.Fields{
					"chain":     wallet.Chain,
					"wallet_id": wallet.ID,
				}).Debug("wallet balance updated")
			}

			if failedCount > 0 {
				log.WithFields(logrus.Fields{
					"chain":  chain,
					"total":  len(group),
					"failed": failedCount,
				}).Warn("chain synchronization finished with errors")
			}
			return nil
		})
	}
	_ = g.Wait()

	log.WithFields(logrus.Fields{
		"total_wallets": len(wallets),
		"duration":      time.Since(start),
	}).Info("wallet synchronization batch completed")

	return nil
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
