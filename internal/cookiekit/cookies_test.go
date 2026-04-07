package cookiekit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldReturnDefaultCookieNames(t *testing.T) {
	names := DefaultCookieNames()

	assert.Equal(t, DefaultAppSessionCookieName, names.AppSession)
	assert.Equal(t, DefaultTwoFactorCookieName, names.TwoFactor)
	assert.Equal(t, DefaultIDPSessionCookieName, names.IDPSession)
}

func TestShouldNormalizeCookieNames(t *testing.T) {
	names := NormalizeCookieNames(CookieNames{
		AppSession: "custom_app_token",
	})

	assert.Equal(t, "custom_app_token", names.AppSession)
	assert.Equal(t, DefaultTwoFactorCookieName, names.TwoFactor)
	assert.Equal(t, DefaultIDPSessionCookieName, names.IDPSession)
}

func TestShouldResolveCookieNamesFromArbitraryValue(t *testing.T) {
	custom := ResolveCookieNames(CookieNames{
		AppSession: "custom_app_token",
		TwoFactor:  "custom_2fa_token",
		IDPSession: "custom_idp_token",
	})
	defaults := ResolveCookieNames("not-cookie-names")

	assert.Equal(t, "custom_app_token", custom.AppSession)
	assert.Equal(t, "custom_2fa_token", custom.TwoFactor)
	assert.Equal(t, "custom_idp_token", custom.IDPSession)
	assert.Equal(t, DefaultCookieNames(), defaults)
}
