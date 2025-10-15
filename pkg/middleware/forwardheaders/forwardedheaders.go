package forwardheaders

import (
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// ---- Functional Options ----

// ForwardHeadersOption is a function type for configuring forwarded headers middleware.
type ForwardHeadersOption func(*ForwardHeadersOptions)

// WithTrustAll configures the middleware to trust forwarded headers from any source.
// WARNING: This is insecure and should only be used in development or when behind a
// trusted reverse proxy that strips untrusted headers.
func WithTrustAll() ForwardHeadersOption {
	return func(o *ForwardHeadersOptions) {
		o.TrustAll = true
	}
}

// WithTrustedProxies sets the list of CIDR ranges or IPs whose forwarded headers are trusted.
// Example: WithTrustedProxies("10.0.0.0/8", "172.16.0.0/12", "192.168.1.100")
func WithTrustedProxies(proxies ...string) ForwardHeadersOption {
	return func(o *ForwardHeadersOptions) {
		o.TrustAll = false
		o.TrustedProxies = proxies
	}
}

// WithRespectForwarded enables parsing RFC 7239 Forwarded header.
// When false, only X-Forwarded-* headers are considered.
func WithRespectForwarded(respect bool) ForwardHeadersOption {
	return func(o *ForwardHeadersOptions) {
		o.RespectForwarded = respect
	}
}

// ---- Options ----

// ForwardHeadersOptions controls how forwarded headers are interpreted.
//
// Security note: trusting forwarded headers from untrusted sources allows
// clients to spoof their origin. Prefer specifying TrustedProxies.
type ForwardHeadersOptions struct {
	// TrustAll accepts forwarded headers from any remote address.
	// Default true to preserve backward compatibility with previous behavior.
	// WARNING: Insecure - only use in development or behind trusted proxy.
	TrustAll bool
	// TrustedProxies is a list of CIDR ranges or IPs whose forwarded headers are trusted.
	// Only used when TrustAll == false.
	TrustedProxies []string
	// RespectForwarded enables parsing RFC 7239 Forwarded header.
	// When false, only X-Forwarded-* headers are considered.
	RespectForwarded bool
}

// ---- Middleware ----

// ---- Middleware ----

// forwardedHeadersMiddleware processes X-Forwarded-* and Forwarded headers to restore original client information.
type forwardedHeadersMiddleware struct {
	opts    ForwardHeadersOptions
	trusted []*net.IPNet
}

// newForwardedHeadersMiddleware builds the middleware and pre-parses trusted proxies.
func newForwardedHeadersMiddleware(opts ForwardHeadersOptions) *forwardedHeadersMiddleware {
	// defaults
	// If TrustAll is false and no TrustedProxies were provided, leave trusted list empty.
	if !opts.RespectForwarded {
		// default to true if unset (zero value is false) to be generous
		opts.RespectForwarded = true
	}
	if opts.TrustAll {
		return &forwardedHeadersMiddleware{opts: opts}
	}
	var nets []*net.IPNet
	for _, s := range opts.TrustedProxies {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if ip := net.ParseIP(s); ip != nil {
			// convert to /32 or /128
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			_, n, _ := net.ParseCIDR(ip.String() + "/" + strconv.Itoa(bits))
			nets = append(nets, n)
			continue
		}
		if _, n, err := net.ParseCIDR(s); err == nil {
			nets = append(nets, n)
		}
	}
	return &forwardedHeadersMiddleware{opts: opts, trusted: nets}
}

// isTrusted returns true when TrustAll is set or the ip is within a trusted range.
func (m *forwardedHeadersMiddleware) isTrusted(ip net.IP) bool {
	if m.opts.TrustAll {
		return true
	}
	if ip == nil {
		return false
	}
	for _, n := range m.trusted {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// firstCSV returns the first comma-separated token, trimmed.
func firstCSV(v string) string {
	if v == "" {
		return ""
	}
	if i := strings.IndexByte(v, ','); i >= 0 {
		v = v[:i]
	}
	return strings.TrimSpace(v)
}

// parseForwardedRFC parses an RFC 7239 Forwarded header entry and extracts for/proto/host.
// The header can be a list; we care about the first element (original client) for host/proto,
// and the left-most for= as the client ip.
func parseForwardedRFC(v string) (forAddr, proto, host string) {
	if v == "" {
		return "", "", ""
	}

	// take first list-element up to comma
	part := v
	if i := strings.IndexByte(v, ','); i >= 0 {
		part = v[:i]
	}

	// split into semicolon-separated key=value pairs
	pairs := strings.Split(part, ";")
	for _, p := range pairs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// split on first '='
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := unquoteAndTrim(kv[1])

		switch {
		case strings.EqualFold(key, "for"):
			forAddr = normalizeForToken(val)
		case strings.EqualFold(key, "proto"):
			proto = val
		case strings.EqualFold(key, "host"):
			host = val
		}
	}

	return
}

// unquoteAndTrim trims whitespace and removes surrounding quotes from a header token.
func unquoteAndTrim(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		return v[1 : len(v)-1]
	}
	return v
}

// normalizeForToken converts a forwarded 'for' token to a plain IP/host without brackets or port.
// It mirrors the previous inlined logic but isolates it for clarity.
func normalizeForToken(v string) string {
	vv := v
	if strings.HasPrefix(vv, "[") {
		if idx := strings.Index(vv, "]"); idx >= 0 {
			vv = vv[1:idx]
		}
	}
	if c := strings.LastIndexByte(vv, ':'); c >= 0 {
		// treat single-colon as host:port; avoid stripping IPv6
		if fc := strings.IndexByte(vv, ':'); fc == c {
			vv = vv[:c]
		}
	}
	return vv
}

// chooseClientIP determines the effective client IP given X-Forwarded-For values and trust list.
// Note: legacy chooseClientIP (operating on pre-split slices) was removed
// in favor of chooseClientIPFromRaw which operates on the raw header string
// and avoids an allocation. The logic is preserved in the Raw variant.

// chooseClientIPFromRaw performs right-to-left parsing of a CSV without building slices.
func (m *forwardedHeadersMiddleware) chooseClientIPFromRaw(xffRaw, immediateRemote string) string {
	if xffRaw == "" {
		return ""
	}
	if m.opts.TrustAll {
		return firstCSV(xffRaw)
	}
	host, _, err := net.SplitHostPort(immediateRemote)
	if err != nil {
		host = immediateRemote
	}
	if !m.isTrusted(net.ParseIP(host)) {
		return ""
	}
	// iterate segments right-to-left
	end := len(xffRaw)
	for end > 0 {
		start := strings.LastIndexByte(xffRaw[:end], ',')
		seg := strings.TrimSpace(xffRaw[start+1 : end])
		v := normalizeIPToken(seg)
		if !m.isTrusted(net.ParseIP(v)) {
			return v
		}
		if start < 0 {
			break
		}
		end = start
	}
	// all trusted; return left-most
	return firstCSV(xffRaw)
}

// normalizeIPToken strips brackets and ports from an IP token.
func normalizeIPToken(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return v
	}
	if strings.HasPrefix(v, "[") {
		if idx := strings.Index(v, "]"); idx >= 0 {
			v = v[1:idx]
		}
	} else {
		if j := strings.LastIndexByte(v, ':'); j >= 0 && strings.Count(v, ":") == 1 {
			v = v[:j]
		}
	}
	return v
}

// Invoke implements the Middleware interface, processing forwarded headers.
func (m *forwardedHeadersMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	r := c.Request()
	hdr := r.Header

	// Fast-path: if no relevant headers are present, skip work.
	fwd := hdr.Get(common.HeaderForwarded)
	xproto := hdr.Get(common.HeaderXForwardedProto)
	xhost := hdr.Get(common.HeaderXForwardedHost)
	xport := hdr.Get(common.HeaderXForwardedPort)
	xffRaw := hdr.Get(common.HeaderXForwardedFor)
	xreal := hdr.Get(common.HeaderXRealIP)
	if fwd == "" && xproto == "" && xhost == "" && xport == "" && xffRaw == "" && xreal == "" {
		next(c)
		return
	}

	// Back-compat: if constructed without options (zero-value), behave permissively
	if !m.opts.TrustAll && !m.opts.RespectForwarded && len(m.trusted) == 0 && len(m.opts.TrustedProxies) == 0 {
		m.opts.TrustAll = true
		m.opts.RespectForwarded = true
	}

	// Trust-gate: if we don't trust headers from this sender, skip parsing entirely
	if !m.shouldApplyHeaders(r) {
		next(c)
		return
	}

	// Extract effective proto, host and client IP (delegated to reduce complexity)
	proto, host, clientIP := m.extractForwarded(r, fwd, xproto, xhost, xport, xffRaw, xreal)

	// Apply changes (we've already trust-gated above)
	if proto != "" {
		r.URL.Scheme = proto
		if proto == "https" && r.TLS == nil {
			// best-effort: mark as TLS by setting a non-nil struct to signal HTTPS
			r.TLS = &tls.ConnectionState{}
		}
	}
	if host != "" {
		r.Host = host
		r.URL.Host = host
		// ensure Host header aligns for downstream usage
		r.Header.Set(common.HeaderHost, host)
	}
	if clientIP != "" {
		// Keep RemoteAddr as IP only (tests assume), drop port
		r.RemoteAddr = clientIP
	}

	next(c)
}

// shouldApplyHeaders determines whether forwarded headers from the request should be trusted.
func (m *forwardedHeadersMiddleware) shouldApplyHeaders(r *http.Request) bool {
	if m.opts.TrustAll {
		return true
	}
	hostPart, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		hostPart = r.RemoteAddr
	}
	return m.isTrusted(net.ParseIP(hostPart))
}

// extractForwarded centralizes header parsing and fallback logic, returning proto, host and clientIP.
func (m *forwardedHeadersMiddleware) extractForwarded(r *http.Request, fwd, xproto, xhost, xport, xffRaw, xreal string) (proto, host, clientIP string) {
	if m.opts.RespectForwarded && fwd != "" {
		fip, p, h := parseForwardedRFC(fwd)
		clientIP, proto, host = fip, p, h
	}
	// Apply X-Forwarded-* fallbacks and determine client IP from headers
	proto, host = updateProtoHostFromX(proto, host, xproto, xhost, xport)
	clientIP = determineClientIPFromHeaders(m, r, clientIP, xffRaw, xreal)
	return
}

// updateProtoHostFromX fills empty proto/host from X-Forwarded-* headers and applies port if needed.
func updateProtoHostFromX(proto, host, xproto, xhost, xport string) (string, string) {
	if proto == "" && xproto != "" {
		proto = firstCSV(xproto)
	}
	if host == "" && xhost != "" {
		host = firstCSV(xhost)
	}
	if xport != "" {
		port := firstCSV(xport)
		if host != "" && !strings.Contains(host, ":") {
			host = host + ":" + port
		}
	}
	return proto, host
}

// determineClientIPFromHeaders chooses client IP from RFC value (if any), X-Forwarded-For, or X-Real-IP.
func determineClientIPFromHeaders(m *forwardedHeadersMiddleware, r *http.Request, current, xffRaw, xreal string) string {
	if current != "" {
		return current
	}
	if xffRaw != "" {
		if chosen := m.chooseClientIPFromRaw(xffRaw, r.RemoteAddr); chosen != "" {
			return chosen
		}
	}
	if xreal != "" {
		return firstCSV(xreal)
	}
	return ""
}

// UseForwardedHeaders adds middleware that processes forwarded headers with the given options.
// By default, TrustAll is enabled for backward compatibility. For security, use
// WithTrustedProxies to specify which proxies to trust.
func UseForwardedHeaders(rtr *router.Router, opts ...ForwardHeadersOption) {
	options := &ForwardHeadersOptions{
		TrustAll:         true, // default for backward compatibility
		RespectForwarded: true,
	}
	for _, opt := range opts {
		opt(options)
	}
	rtr.Use(newForwardedHeadersMiddleware(*options))
}
