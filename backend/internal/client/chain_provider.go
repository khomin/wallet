package client

import (
	"context"
	"time"
	"tracker/config"
	"tracker/internal/metrics"

	"golang.org/x/time/rate"
)

type ChainProvider interface {
	GetBalance(ctx context.Context, address string) (float64, error)
	GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error)
	ValidateAddress(address, tokenAddress string) error
	Connect(ctx context.Context) error
	Close()
}

type RateLimitedProvider struct {
	ChainProvider
	limiter *rate.Limiter
	chain   string
}

func NewRateLimiterProvider(
	cfg *config.ChainRateLimitConfig,
	provider ChainProvider,
	chain string,
) *RateLimitedProvider {
	return &RateLimitedProvider{
		limiter:       rate.NewLimiter(rate.Limit(cfg.RPS), cfg.Burst),
		chain:         chain,
		ChainProvider: provider,
	}
}

func (r *RateLimitedProvider) GetBalance(ctx context.Context, address string) (float64, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return 0, err
	}
	start := time.Now()
	balance, err := r.ChainProvider.GetBalance(ctx, address)

	metrics.BlockchainRequests.
		WithLabelValues(r.chain, "get_balance", r.status(err)).
		Inc()

	metrics.BlockchainRequestDuration.
		WithLabelValues(r.chain, "get_balance").
		Observe(time.Since(start).Seconds())

	return balance, err
}

func (r *RateLimitedProvider) GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return 0, err
	}
	start := time.Now()
	balance, err := r.ChainProvider.GetTokenBalance(ctx, address, tokenAddress)

	metrics.BlockchainRequests.
		WithLabelValues(r.chain, "get_token_balance", r.status(err)).
		Inc()

	metrics.BlockchainRequestDuration.
		WithLabelValues(r.chain, "get_token_balance").
		Observe(time.Since(start).Seconds())

	return balance, err
}

func (r *RateLimitedProvider) status(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}
