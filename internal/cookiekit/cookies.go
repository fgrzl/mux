package cookiekit

import internalcommon "github.com/fgrzl/mux/internal/common"

const (
	DefaultAppSessionCookieName = "app_token"
	DefaultTwoFactorCookieName  = "2fa_token"
	DefaultIDPSessionCookieName = "idp_token"
)

const ServiceKeyCookieNames = internalcommon.ServiceKey("mux.auth.cookie_names")

type CookieNames struct {
	AppSession string
	TwoFactor  string
	IDPSession string
}

func DefaultCookieNames() CookieNames {
	return CookieNames{
		AppSession: DefaultAppSessionCookieName,
		TwoFactor:  DefaultTwoFactorCookieName,
		IDPSession: DefaultIDPSessionCookieName,
	}
}

func GetUserCookieName() string {
	return DefaultAppSessionCookieName
}

func GetTwoFactorCookieName() string {
	return DefaultTwoFactorCookieName
}

func GetIdpSessionCookieName() string {
	return DefaultIDPSessionCookieName
}

func NormalizeCookieNames(names CookieNames) CookieNames {
	defaults := DefaultCookieNames()
	if names.AppSession == "" {
		names.AppSession = defaults.AppSession
	}
	if names.TwoFactor == "" {
		names.TwoFactor = defaults.TwoFactor
	}
	if names.IDPSession == "" {
		names.IDPSession = defaults.IDPSession
	}
	return names
}

func ResolveCookieNames(value any) CookieNames {
	names, ok := value.(CookieNames)
	if !ok {
		return DefaultCookieNames()
	}
	return NormalizeCookieNames(names)
}
