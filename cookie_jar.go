package mux

import "sync"

type CookieJarOptions struct {
	AppSessionCookieName string
	TwoFactorCookieName  string
	IdpSessionCookieName string
}

var (
	appSessionName string
	twoFactorName  string
	idpSessionName string
	once           sync.Once
	rwMu           sync.RWMutex
)

// Define constants for default cookie names
const (
	DefaultAppSessionCookieName = "app_session"
	DefaultTwoFactorCookieName  = "2fa_session"
	DefaultIdpSessionCookieName = "idp_session"
)

// Getter and Setter for AppSessionCookieName
func GetAppSessionCookieName() string {
	rwMu.RLock()
	defer rwMu.RUnlock()
	if appSessionName == "" {
		return DefaultAppSessionCookieName
	}
	return appSessionName
}

func SetAppSessionCookieName(name string) {
	rwMu.Lock()
	defer rwMu.Unlock()
	appSessionName = name
}

// Getter and Setter for TwoFactorCookieName
func GetTwoFactorCookieName() string {
	rwMu.RLock()
	defer rwMu.RUnlock()
	if twoFactorName == "" {
		return DefaultTwoFactorCookieName
	}
	return twoFactorName
}

func SetTwoFactorCookieName(name string) {
	rwMu.Lock()
	defer rwMu.Unlock()
	twoFactorName = name
}

// Getter and Setter for IdpSessionCookieName
func GetIdpSessionCookieName() string {
	rwMu.RLock()
	defer rwMu.RUnlock()
	if idpSessionName == "" {
		return DefaultIdpSessionCookieName
	}
	return idpSessionName
}

func SetIdpSessionCookieName(name string) {
	rwMu.Lock()
	defer rwMu.Unlock()
	idpSessionName = name
}
