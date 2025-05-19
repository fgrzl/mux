package mux

import "sync"

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
	DefaultUserCookieName      = "app_token"
	DefaultTwoFactorCookieName = "2fa_token"
	DefaultIdpUserCookieName   = "idp_token"
)

// Getter and Setter for AppSessionCookieName
func GetUserCookieName() string {
	rwMu.RLock()
	defer rwMu.RUnlock()
	if userCookieName == "" {
		return DefaultUserCookieName
	}
	return userCookieName
}

func SetAppSessionCookieName(name string) {
	rwMu.Lock()
	defer rwMu.Unlock()
	userCookieName = name
}

// Getter and Setter for TwoFactorCookieName
func GetTwoFactorCookieName() string {
	rwMu.RLock()
	defer rwMu.RUnlock()
	if twoFactorCookieName == "" {
		return DefaultTwoFactorCookieName
	}
	return twoFactorCookieName
}

func SetTwoFactorCookieName(name string) {
	rwMu.Lock()
	defer rwMu.Unlock()
	twoFactorCookieName = name
}

// Getter and Setter for IdpSessionCookieName
func GetIdpSessionCookieName() string {
	rwMu.RLock()
	defer rwMu.RUnlock()
	if idpUserCookieName == "" {
		return DefaultIdpUserCookieName
	}
	return idpUserCookieName
}

func SetIdpSessionCookieName(name string) {
	rwMu.Lock()
	defer rwMu.Unlock()
	idpUserCookieName = name
}
