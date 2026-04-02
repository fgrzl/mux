package tokenizer

import (
	"context"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/internal/common"
)

const ServiceKeyTokenProvider common.ServiceKey = "tokenizer.token_provider"

// TokenProvider defines the minimal interface for creating and validating tokens.
// It uses context.Context so internal packages don't need to import routing types.
type TokenProvider interface {
	CreateToken(ctx context.Context, principal claims.Principal) (string, error)
	ValidateToken(ctx context.Context, token string) (claims.Principal, error)
	GetTTL() time.Duration
	CanCreateTokens() bool
}
