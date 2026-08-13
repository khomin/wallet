package client

import (
	"context"
	"tracker/config"

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
}

func NewRateLimiterProvider(
	cfg *config.ChainRateLimitConfig,
	provider ChainProvider,
) *RateLimitedProvider {
	return &RateLimitedProvider{
		limiter:       rate.NewLimiter(rate.Limit(cfg.RPS), cfg.Burst),
		ChainProvider: provider,
	}
}

func (r *RateLimitedProvider) GetBalance(ctx context.Context, address string) (float64, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return 0, err
	}
	return r.ChainProvider.GetBalance(ctx, address)
}

func (r *RateLimitedProvider) GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return 0, err
	}
	return r.ChainProvider.GetTokenBalance(ctx, address, tokenAddress)
}
