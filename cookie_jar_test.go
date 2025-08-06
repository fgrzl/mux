package mux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldReturnDefaultUserCookieNameWhenNotSet(t *testing.T) {
	// Arrange
	SetAppSessionCookieName("") // Reset to default

	// Act
	name := GetUserCookieName()

	// Assert
	assert.Equal(t, DefaultUserCookieName, name)
}

func TestShouldReturnCustomUserCookieNameWhenSet(t *testing.T) {
	// Arrange
	customName := "custom_app_token"
	SetAppSessionCookieName(customName)

	// Act
	name := GetUserCookieName()

	// Assert
	assert.Equal(t, customName, name)

	// Cleanup
	SetAppSessionCookieName("")
}

func TestShouldReturnDefaultTwoFactorCookieNameWhenNotSet(t *testing.T) {
	// Arrange
	SetTwoFactorCookieName("") // Reset to default

	// Act
	name := GetTwoFactorCookieName()

	// Assert
	assert.Equal(t, DefaultTwoFactorCookieName, name)
}

func TestShouldReturnCustomTwoFactorCookieNameWhenSet(t *testing.T) {
	// Arrange
	customName := "custom_2fa_token"
	SetTwoFactorCookieName(customName)

	// Act
	name := GetTwoFactorCookieName()

	// Assert
	assert.Equal(t, customName, name)

	// Cleanup
	SetTwoFactorCookieName("")
}

func TestShouldReturnDefaultIdpSessionCookieNameWhenNotSet(t *testing.T) {
	// Arrange
	SetIdpSessionCookieName("") // Reset to default

	// Act
	name := GetIdpSessionCookieName()

	// Assert
	assert.Equal(t, DefaultIdpUserCookieName, name)
}

func TestShouldReturnCustomIdpSessionCookieNameWhenSet(t *testing.T) {
	// Arrange
	customName := "custom_idp_token"
	SetIdpSessionCookieName(customName)

	// Act
	name := GetIdpSessionCookieName()

	// Assert
	assert.Equal(t, customName, name)

	// Cleanup
	SetIdpSessionCookieName("")
}

func TestShouldHandleConcurrentAccessToCookieNames(t *testing.T) {
	// Arrange
	customName := "concurrent_test_token"

	// Act & Assert - This test mainly ensures no data races occur
	go func() {
		SetAppSessionCookieName(customName)
	}()

	go func() {
		name := GetUserCookieName()
		assert.NotEmpty(t, name)
	}()

	// Wait a bit for goroutines to complete
	name := GetUserCookieName()
	assert.NotEmpty(t, name)

	// Cleanup
	SetAppSessionCookieName("")
}
