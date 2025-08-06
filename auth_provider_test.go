package mux

import (
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/stretchr/testify/assert"
)

// MockSigner implements jwtkit.Signer for testing
type MockSigner struct {
	shouldError bool
}

func (m *MockSigner) CreateToken(principal claims.Principal, ttl time.Duration) (string, error) {
	if m.shouldError {
		return "", assert.AnError
	}
	return "mock-token", nil
}

func TestShouldCreateAuthProviderWithDefaultTTL(t *testing.T) {
	// Arrange
	signer := &MockSigner{}

	// Act
	provider := NewAuthProvider(signer, nil)

	// Assert
	assert.NotNil(t, provider)
	assert.Implements(t, (*AuthProvider)(nil), provider)
}

func TestShouldCreateAuthProviderWithCustomTTL(t *testing.T) {
	// Arrange
	signer := &MockSigner{}
	customTTL := time.Hour

	// Act
	provider := NewAuthProvider(signer, &customTTL)

	// Assert
	assert.NotNil(t, provider)
	assert.Implements(t, (*AuthProvider)(nil), provider)
}