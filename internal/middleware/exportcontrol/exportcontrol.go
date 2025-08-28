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

type ExportControlOptions struct {
	DB *geoip2.Reader
}

type ExportControlOption func(*ExportControlOptions)

func WithGeoIPDatabase(db *geoip2.Reader) ExportControlOption {
	return func(o *ExportControlOptions) {
		o.DB = db
	}
}

func UseExportControl(rtr *router.Router, opts ...ExportControlOption) {
	options := &ExportControlOptions{}
	for _, opt := range opts {
		opt(options)
	}
	// Register middleware using the exported API.
	rtr.Use(&exportControlMiddleware{options: options})
}

// ---- Middleware ----

type exportControlMiddleware struct {
	options *ExportControlOptions
}

func (m *exportControlMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	ip := getRealIP(c.Request())
	if parsed := net.ParseIP(ip); parsed != nil && m.options.DB != nil {
		if record, err := m.options.DB.Country(parsed); err == nil {
			if _, blocked := exportRestrictedCountries[record.Country.IsoCode]; blocked {
				c.Response().WriteHeader(http.StatusForbidden)
				_, _ = c.Response().Write([]byte("Access from your country is restricted due to export control policies."))
				return
			}
		}
	}
	next(c)
}

// ---- Helpers ----

var exportRestrictedCountries = map[string]struct{}{
	"IR": {},
	"KP": {},
	"SY": {},
	"CU": {},
	"RU": {},
}

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
