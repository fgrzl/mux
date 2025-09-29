package test

import (
	"bufio"
	"fmt"
	"net"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/test/testsupport"
	"github.com/stretchr/testify/require"
)

// TestSimpleWebSocketUpgrade exercises the mux router's `/ws` route from testsupport
// by spinning up a real mux.Router and performing a minimal handshake.
func TestSimpleWebSocketUpgrade(t *testing.T) {
	r := router.NewRouter(router.WithHeadFallbackToGet())
	testsupport.ConfigureRoutes(r)

	srv := httptest.NewServer(r)
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	req := "GET /ws HTTP/1.1\r\n"
	req += fmt.Sprintf("Host: %s\r\n", addr)
	req += "Upgrade: websocket\r\n"
	req += "Connection: Upgrade\r\n"
	req += "Sec-WebSocket-Version: 13\r\n"
	// Example key from RFC 6455
	req += "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n"
	req += "\r\n"

	_, err = conn.Write([]byte(req))
	require.NoError(t, err)

	br := bufio.NewReader(conn)
	line, err := br.ReadString('\n')
	require.NoError(t, err)
	require.Contains(t, line, "101")

	// Read headers until blank line
	for {
		h, err := br.ReadString('\n')
		require.NoError(t, err)
		if strings.TrimSpace(h) == "" {
			break
		}
	}
}
