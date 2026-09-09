package core

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	onBalanceChanged  func(context.Context, domain.WalletBalanceChange)
	streamSubscribers map[string]map[string]chan *walletv1.WalletUpdate
	lock              sync.RWMutex
}

type NewWalletDeps struct {
	WalletService    *WalletService
	WalletRepo       WalletRepository
	OnBalanceChanged func(context.Context, domain.WalletBalanceChange)
}

func NewWalletWorker(deps *NewWalletDeps) *WalletWorker {
	return &WalletWorker{
		walletRepo:        deps.WalletRepo,
		walletService:     deps.WalletService,
		onBalanceChanged:  deps.OnBalanceChanged,
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

	wallets, err := w.walletRepo.ListForSync(ctx, time.Now().Add(-5*time.Minute), 100)
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

	for chain, wallets := range groupsByChain {
		chain := chain
		wallets := wallets

		g.Go(func() error {
			var failedCount int

			for _, wallet := range wallets {
				result, err := w.synchronizeBalance(ctx, wallet)
				if err != nil {
					failedCount++
					log.Errorf("failed to sync wallet balance: %s %s", wallet.Chain, wallet.ID)
					continue
				}
				log.Debugf("wallet balance updated: %s %s", wallet.Chain, wallet.ID)

				id, err := uuid.Parse(wallet.ID)
				if err != nil {
					log.Errorf("failed to parse wallet id: %s", wallet.ID)
					continue
				}
				users, err := w.walletRepo.ListUsersByWallet(ctx, id)
				if err != nil {
					log.Errorf("failed to list users for the wallet id: %s", wallet.ID)
					continue
				}
				for _, userWallet := range users {
					streams, ok := w.streamSubscribers[userWallet.UserID]
					if ok {
						for _, v := range streams {
							v <- &walletv1.WalletUpdate{
								Wallet: (&domain.UserWalletBalance{
									UserWallet: userWallet,
									Balance:    result.newBalance.Balance,
									BalanceUSD: result.newBalance.BalanceUSD,
									UpdatedAt:  result.newBalance.Time,
								}).ToGrpc(),
							}
						}
					}
					if result.changed {
						w.onBalanceChanged(ctx, domain.WalletBalanceChange{
							Wallet:            wallet,
							CurrentBalance:    result.newBalance.Balance,
							CurrentBalanceUSD: result.newBalance.BalanceUSD,
							OldBalance:        result.oldBalance.Balance,
							OldBalanceUSD:     result.oldBalance.BalanceUSD,
						})
					}
				}
			}
			if failedCount > 0 {
				log.Warnf("chain synchronization finished with errors: %s %d, %d", chain, len(wallets), failedCount)
			}
			return nil
		})
	}
	_ = g.Wait()

	log.Infof("wallet synchronization batch completed: %d, %v", len(wallets), time.Since(start))

	return nil
}

func (w *WalletWorker) synchronizeBalance(ctx context.Context, wallet domain.Wallet) (*updateBalanceResult, error) {
	uuid, err := uuid.Parse(wallet.ID)
	if err != nil {
		return nil, err
	}
	newBalance, err := w.walletService.FetchBalance(ctx, wallet)
	if err != nil {
		if errors.Is(err, ErrProviderTimeout) {
			return nil, err
		}
		if errors.Is(err, ErrProviderRateLimit) {
			return nil, err
		}
		return nil, err
	}
	oldBalance, err := w.walletRepo.GetBalanceSnapshot(ctx, uuid)
	if err != nil {
		if !errors.Is(err, domain.ErrorNotFound) {
			return nil, err
		}
	}
	err = w.walletRepo.UpdateBalanceSnapshot(ctx, uuid, BalanceSnapshot{
		Crypto: newBalance.Balance,
		USD:    newBalance.BalanceUSD,
		Time:   time.Now(),
	})
	if err != nil {
		return nil, err
	}
	if oldBalance != nil && !oldBalance.Time.IsZero() {
		delta := newBalance.Balance - oldBalance.Balance
		if math.Abs(delta) >= 0.000001 {
			return &updateBalanceResult{
				newBalance: newBalance,
				oldBalance: oldBalance,
				changed:    true,
			}, nil
		}
	}
	return &updateBalanceResult{
		newBalance: newBalance,
		changed:    false,
	}, nil
}

type updateBalanceResult struct {
	newBalance *domain.WalletBalanceSnapshot
	oldBalance *domain.WalletBalanceSnapshot
	changed    bool
}
