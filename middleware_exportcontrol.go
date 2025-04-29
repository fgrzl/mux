package mux

import (
	"net"
	"net/http"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// List of ISO country codes to block under export control restrictions
var exportRestrictedCountries = map[string]struct{}{
	"IR": {}, // Iran
	"KP": {}, // North Korea
	"SY": {}, // Syria
	"CU": {}, // Cuba
	"RU": {}, // Russia (optional, based on policy)
}

// ExportControlOptions holds config for the middleware
type ExportControlOptions struct {
	DB *geoip2.Reader
}

// UseExportControl attaches the export control middleware to the router
func (rtr *Router) UseExportControl(options *ExportControlOptions) {
	rtr.middleware = append(rtr.middleware, &exportControlMiddleware{options: options})
}

type exportControlMiddleware struct {
	options *ExportControlOptions
}

// Invoke implements the middleware logic for export control checks
func (m *exportControlMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	ip := getRealIP(c.Request)
	if parsed := net.ParseIP(ip); parsed != nil && m.options.DB != nil {
		if record, err := m.options.DB.Country(parsed); err == nil {
			if _, blocked := exportRestrictedCountries[record.Country.IsoCode]; blocked {
				c.Response.WriteHeader(http.StatusForbidden)
				_, _ = c.Response.Write([]byte("Access from your country is restricted due to export control policies."))
				return
			}
		}
	}
	next(c)
}

func getRealIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
