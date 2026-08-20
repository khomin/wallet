package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	walletv1 "tracker/gen/wallet/v1"
	"tracker/internal/core/domain"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type WalletWorker struct {
	walletService     *WalletService
	walletRepo        WalletRepository
	streamSubscribers map[string]map[string]chan *walletv1.WalletUpdate
	lock              sync.RWMutex
}

type NewWalletDeps struct {
	WalletService *WalletService
	WalletRepo    WalletRepository
}

func NewWalletWorker(deps *NewWalletDeps) *WalletWorker {
	return &WalletWorker{
		walletRepo:        deps.WalletRepo,
		walletService:     deps.WalletService,
		streamSubscribers: make(map[string]map[string]chan *walletv1.WalletUpdate),
		lock:              sync.RWMutex{},
	}
}

func (w *WalletWorker) Start(ctx context.Context) {
	go w.startSyncLoop(ctx, time.Second*1)
}

func (w *WalletWorker) Subscribe(userID string) (<-chan *walletv1.WalletUpdate, string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	streamID := uuid.NewString()
	channel := make(chan *walletv1.WalletUpdate)
	if w.streamSubscribers[userID][streamID] == nil {
		w.streamSubscribers[userID] = make(map[string]chan *walletv1.WalletUpdate)
	}
	w.streamSubscribers[userID][streamID] = channel
	return channel, streamID
}

func (w *WalletWorker) UnSubscribe(userID string, streamID string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	subs, exists := w.streamSubscribers[userID]
	if !exists {
		return
	}
	channel, exists := subs[streamID]
	if !exists {
		return
	}
	close(channel)
	delete(subs, streamID)
}

func (w *WalletWorker) startSyncLoop(ctx context.Context, interval time.Duration) {
	tm := time.NewTicker(interval)
	defer tm.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.C:
			w.synchronizeWallets(ctx)
		}
	}
}

func (w *WalletWorker) synchronizeWallets(ctx context.Context) error {
	start := time.Now()
	g, ctx := errgroup.WithContext(ctx)
	log := logrus.WithContext(ctx)
	groupsByChain := make(map[string][]domain.Wallet)

	wallets, err := w.walletRepo.ListForSync(ctx, 100)
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

	for chain, group := range groupsByChain {
		chain := chain
		group := group

		g.Go(func() error {
			var failedCount int

			for _, wallet := range group {
				balance, err := w.updateBalance(ctx, wallet)
				if err != nil {
					failedCount++
					log.Errorf("failed to sync wallet balance: %s %s", wallet.Chain, wallet.ID)
					continue
				}
				log.Debugf("wallet balance updated: %s %s", wallet.Chain, wallet.ID)

				streams, ok := w.streamSubscribers[wallet.UserID]
				if ok {
					for _, v := range streams {
						v <- &walletv1.WalletUpdate{
							Wallet: balance.ToGrpc(),
						}
					}
				}
			}
			if failedCount > 0 {
				log.Warnf("chain synchronization finished with errors: %s %d, %d", chain, len(group), failedCount)
			}
			return nil
		})
	}
	_ = g.Wait()

	log.Infof("wallet synchronization batch completed: %d, %v", len(wallets), time.Since(start))

	return nil
}

func (w *WalletWorker) updateBalance(ctx context.Context, wallet domain.Wallet) (*domain.WalletWithBalance, error) {
	uuid, err := uuid.Parse(wallet.ID)
	if err != nil {
		return nil, err
	}
	balance, err := w.walletService.FetchBalance(ctx, wallet)
	if err != nil {
		if errors.Is(err, ErrProviderTimeout) {
			return nil, err
		}
		if errors.Is(err, ErrProviderRateLimit) {
			return nil, err
		}
		return nil, err
	}
	err = w.walletRepo.UpdateBalance(ctx, wallet.UserID, uuid, balance.Balance, balance.BalanceUSD)
	if err != nil {
		return nil, err
	}
	return balance, nil
}
