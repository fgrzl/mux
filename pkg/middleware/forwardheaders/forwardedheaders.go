package forwardheaders

import (
	"crypto/tls"
	"net"
	"strconv"
	"strings"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// Options controls how forwarded headers are interpreted.
//
// Security note: trusting forwarded headers from untrusted sources allows
// clients to spoof their origin. Prefer specifying TrustedProxies.
type Options struct {
	// TrustAll accepts forwarded headers from any remote address.
	// Default true to preserve backward compatibility with previous behavior.
	TrustAll bool
	// TrustedProxies is a list of CIDR ranges or IPs whose forwarded headers are trusted.
	// Only used when TrustAll == false.
	TrustedProxies []string
	// RespectForwarded enables parsing RFC 7239 Forwarded header.
	// When false, only X-Forwarded-* headers are considered.
	RespectForwarded bool
}

// forwardedHeadersMiddleware processes X-Forwarded-* and Forwarded headers to restore original client information.
type forwardedHeadersMiddleware struct {
	opts    Options
	trusted []*net.IPNet
}

// newForwardedHeadersMiddleware builds the middleware and pre-parses trusted proxies.
func newForwardedHeadersMiddleware(opts Options) *forwardedHeadersMiddleware {
	// defaults
	if opts.TrustAll == false && len(opts.TrustedProxies) == 0 {
		// remain explicit: TrustAll false means only exact trust list; leave empty
	}
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

// splitCSV splits a comma-separated header into trimmed non-empty parts.
func splitCSV(v string) []string { // kept for compatibility if used by tests; not used in hot path
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
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
	end := len(v)
	if i := strings.IndexByte(v, ','); i >= 0 {
		end = i
	}
	i := 0
	// iterate semicolon-separated kv pairs
	for i < end {
		// skip leading spaces/semicolons
		for i < end && (v[i] == ' ' || v[i] == ';') {
			i++
		}
		if i >= end {
			break
		}
		// key up to '='
		ks := i
		for i < end && v[i] != '=' && v[i] != ';' {
			i++
		}
		if i >= end || v[i] != '=' { // malformed or empty
			// skip to next ';'
			for i < end && v[i] != ';' {
				i++
			}
			continue
		}
		ke := i
		i++ // skip '='
		// value up to ';' or end, trim spaces
		vs := i
		// handle quoted value
		var val string
		if vs < end && v[vs] == '"' {
			// quoted string: find next '"'
			vs++
			j := vs
			for j < end && v[j] != '"' {
				j++
			}
			val = v[vs:j]
			i = j + 1
		} else {
			// unquoted until ';' or end
			j := vs
			for j < end && v[j] != ';' {
				j++
			}
			// trim spaces on both ends
			// left trim
			for vs < j && v[vs] == ' ' {
				vs++
			}
			je := j
			for je > vs && v[je-1] == ' ' {
				je--
			}
			val = v[vs:je]
			i = j
		}
		// normalize key trim spaces
		// left trim
		for ks < ke && v[ks] == ' ' {
			ks++
		}
		for ke > ks && v[ke-1] == ' ' {
			ke--
		}
		key := v[ks:ke]
		if strings.EqualFold(key, "for") {
			vv := val
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
			forAddr = vv
		} else if strings.EqualFold(key, "proto") {
			proto = val
		} else if strings.EqualFold(key, "host") {
			host = val
		}
		// move to next after optional ';'
		if i < end && v[i] == ';' {
			i++
		}
	}
	return
}

// chooseClientIP determines the effective client IP given X-Forwarded-For values and trust list.
func (m *forwardedHeadersMiddleware) chooseClientIP(xff []string, immediateRemote string) string { // legacy path
	if len(xff) == 0 {
		return ""
	}
	if m.opts.TrustAll {
		return xff[0]
	}
	host, _, err := net.SplitHostPort(immediateRemote)
	if err != nil {
		host = immediateRemote
	}
	if !m.isTrusted(net.ParseIP(host)) {
		return ""
	}
	for i := len(xff) - 1; i >= 0; i-- {
		v := normalizeIPToken(xff[i])
		if !m.isTrusted(net.ParseIP(v)) {
			return v
		}
	}
	return xff[0]
}

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
	// Back-compat: if constructed without options (zero-value), behave permissively
	if !m.opts.TrustAll && !m.opts.RespectForwarded && len(m.trusted) == 0 && len(m.opts.TrustedProxies) == 0 {
		m.opts.TrustAll = true
		m.opts.RespectForwarded = true
	}
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

	// Prefer RFC 7239 Forwarded when enabled
	var proto, host, clientIP string

	// Trust-gate: if we don't trust headers from this sender, skip parsing entirely
	applyHeaders := m.opts.TrustAll
	if !applyHeaders {
		hostPart, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			hostPart = r.RemoteAddr
		}
		// parse once
		applyHeaders = m.isTrusted(net.ParseIP(hostPart))
	}
	if !applyHeaders {
		next(c)
		return
	}
	if m.opts.RespectForwarded {
		if fwd != "" {
			fip, p, h := parseForwardedRFC(fwd)
			clientIP, proto, host = fip, p, h
		}
	}

	// Fallback/augment from X-Forwarded-* headers
	if proto == "" {
		if xproto != "" {
			proto = firstCSV(xproto)
		}
	}
	// Host: prefer X-Forwarded-Host
	if host == "" {
		if xhost != "" {
			host = firstCSV(xhost)
		}
	}
	// If port provided separately, combine
	if xport != "" {
		port := firstCSV(xport)
		if host != "" && !strings.Contains(host, ":") {
			host = host + ":" + port
		}
	}

	// Determine client IP from X-Forwarded-For or X-Real-IP if not from RFC header
	if clientIP == "" {
		if xffRaw != "" {
			if chosen := m.chooseClientIPFromRaw(xffRaw, r.RemoteAddr); chosen != "" {
				clientIP = chosen
			}
		}
		if clientIP == "" && xreal != "" {
			clientIP = firstCSV(xreal)
		}
	}

	// Apply changes (we've already trust-gated above)
	if applyHeaders {
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
			r.Header.Set("Host", host)
		}
	}
	if clientIP != "" && applyHeaders {
		// Keep RemoteAddr as IP only (tests assume), drop port
		r.RemoteAddr = clientIP
	}

	next(c)
}

// UseForwardedHeaders adds middleware that processes forwarded headers with permissive defaults.
func UseForwardedHeaders(rtr *router.Router) {
	rtr.Use(newForwardedHeadersMiddleware(Options{TrustAll: true, RespectForwarded: true}))
}

// UseForwardedHeadersWithOptions adds middleware with custom options.
func UseForwardedHeadersWithOptions(rtr *router.Router, opts Options) {
	rtr.Use(newForwardedHeadersMiddleware(opts))
}
