package mux

import "sync"

// CookieJarOptions configures cookie names for different types of authentication cookies.
type CookieJarOptions struct {
	AppSessionCookieName string
	TwoFactorCookieName  string
	IdpSessionCookieName string
}

var (
	userCookieName      string
	twoFactorCookieName string
	idpUserCookieName   string
	rwMu                sync.RWMutex
)

// Define constants for default cookie names
const (
	// DefaultUserCookieName is the default name for the application session cookie.
	DefaultUserCookieName = "app_token"
	// DefaultTwoFactorCookieName is the default name for the two-factor authentication cookie.
	DefaultTwoFactorCookieName = "2fa_token"
	// DefaultIdpUserCookieName is the default name for the identity provider session cookie.
	DefaultIdpUserCookieName = "idp_token"
)

// GetUserCookieName returns the current application session cookie name.
// If not set, returns the default name.
func GetUserCookieName() string {
	rwMu.RLock()
	defer rwMu.RUnlock()
	if userCookieName == "" {
		return DefaultUserCookieName
	}
	return userCookieName
}

// SetAppSessionCookieName sets the application session cookie name.
func SetAppSessionCookieName(name string) {
	rwMu.Lock()
	defer rwMu.Unlock()
	userCookieName = name
}

// GetTwoFactorCookieName returns the current two-factor authentication cookie name.
// If not set, returns the default name.
func GetTwoFactorCookieName() string {
	rwMu.RLock()
	defer rwMu.RUnlock()
	if twoFactorCookieName == "" {
		return DefaultTwoFactorCookieName
	}
	return twoFactorCookieName
}

// SetTwoFactorCookieName sets the two-factor authentication cookie name.
func SetTwoFactorCookieName(name string) {
	rwMu.Lock()
	defer rwMu.Unlock()
	twoFactorCookieName = name
}

// GetIdpSessionCookieName returns the current identity provider session cookie name.
// If not set, returns the default name.
func GetIdpSessionCookieName() string {
	rwMu.RLock()
	defer rwMu.RUnlock()
	if idpUserCookieName == "" {
		return DefaultIdpUserCookieName
	}
	return idpUserCookieName
}

// SetIdpSessionCookieName sets the identity provider session cookie name.
func SetIdpSessionCookieName(name string) {
	rwMu.Lock()
	defer rwMu.Unlock()
	idpUserCookieName = name
}
