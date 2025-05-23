package mux

import (
	"context"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/claims/jwtkit"
)

// AuthProvider is the interface that wraps token creation.
type AuthProvider interface {
	CreateToken(ctx context.Context, principal claims.Principal) (string, error)
}

type authProvider struct {
	signer jwtkit.Signer
	ttl    time.Duration
}

// NewAuthProvider creates a new AuthProvider using the given signer and optional TTL.
func NewAuthProvider(signer jwtkit.Signer, ttl *time.Duration) AuthProvider {
	finalTTL := time.Minute * 15
	if ttl != nil {
		finalTTL = *ttl
	}
	return &authProvider{
		signer: signer,
		ttl:    finalTTL,
	}
}

// CreateToken generates a signed JWT for the given principal.
func (a *authProvider) CreateToken(ctx context.Context, principal claims.Principal) (string, error) {
	return a.signer.CreateToken(principal, a.ttl)
}
