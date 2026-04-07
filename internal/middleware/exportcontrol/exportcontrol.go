package exportcontrol

import (
	"net"
	"net/http"
	"strings"

	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
	"github.com/oschwald/geoip2-golang"
)

// ---- Functional Options ----

// ExportControlOptions configures the export control middleware.
type ExportControlOptions struct {
	DB *geoip2.Reader
}

// ExportControlOption configures ExportControlOptions via functional options.
type ExportControlOption func(*ExportControlOptions)

// WithGeoIPDatabase sets the GeoIP database to use when determining the request's country.
func WithGeoIPDatabase(db *geoip2.Reader) ExportControlOption {
	return func(o *ExportControlOptions) {
		o.DB = db
	}
}

// UseExportControl adds middleware that denies requests originating from restricted countries.
func UseExportControl(rtr *router.Router, opts ...ExportControlOption) {
	options := &ExportControlOptions{}
	for _, opt := range opts {
		opt(options)
	}
	// Register middleware using the exported API.
	rtr.Use(&exportControlMiddleware{options: options})
}

// ---- Middleware ----

// exportControlMiddleware enforces export control restrictions based on GeoIP lookup.
type exportControlMiddleware struct {
	options *ExportControlOptions
}

const restrictedMessage = "Access from your country is restricted due to export control policies."

// Invoke implements the Middleware interface, denying access when the client IP resolves to a restricted country.
func (m *exportControlMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	ip := getRealIP(c.Request())
	if ip != "" && m.options.DB != nil {
		if parsed := net.ParseIP(ip); parsed != nil {
			if record, err := m.options.DB.Country(parsed); err == nil {
				if _, blocked := exportRestrictedCountries[record.Country.IsoCode]; blocked {
					c.Forbidden(restrictedMessage)
					return
				}
			}
		}
	}
	next(c)
}

// ---- Helpers ----

// exportRestrictedCountries is the set of ISO country codes that are not allowed.
var exportRestrictedCountries = map[string]struct{}{
	"IR": {},
	"KP": {},
	"SY": {},
	"CU": {},
	"RU": {},
}

// getRealIP attempts to determine the client's IP address from headers or the remote address.
func getRealIP(r *http.Request) string {
	if xff := r.Header.Get(common.HeaderXForwardedFor); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if xrip := r.Header.Get(common.HeaderXRealIP); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
