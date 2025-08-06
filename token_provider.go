package mux

import (
	"time"

	"github.com/fgrzl/claims"
)

type TokenProvider interface {
	CreateToken(ctx *RouteContext, principal claims.Principal) (string, error)
	ValidateToken(ctx *RouteContext, token string) (claims.Principal, error)
	GetTTL() time.Duration
	CanCreateTokens() bool
}
