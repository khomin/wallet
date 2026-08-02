package middleware

import (
	"context"
	"fmt"
	"tracker/internal/core/domain"

	"github.com/coreos/go-oidc/v3/oidc"
)

type tokenVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func NewTokenVerifier(ctx context.Context, issuerURL string, clientID string) (TokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}
	// Configure the verifier with your client ID
	// This will validate the 'aud' claim matches your client
	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})
	return &tokenVerifier{
		verifier: verifier,
	}, nil
}

func (v *tokenVerifier) Verify(ctx context.Context, rawToken string) (domain.JwtClaims, error) {
	return domain.JwtClaims{}, nil
}
