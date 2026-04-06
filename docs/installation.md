# Installation

This guide covers prerequisites, module setup, verification, and local development for Mux.

## Requirements

- Go 1.25.6 or later
- Linux, macOS, or Windows
- Go modules enabled

Check your Go version:

```bash
go version
```

## Add Mux to an Existing Project

```bash
go get github.com/fgrzl/mux
go mod tidy
```

## Create a New Project

```bash
mkdir my-api
cd my-api
go mod init my-api
go get github.com/fgrzl/mux
```

## Verify Your Installation

Create `main.go`:

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()

    if err := router.Configure(func(router *mux.Router) {
        router.GET("/", func(c mux.RouteContext) {
            c.OK("Mux is working!")
        })
    }); err != nil {
        panic(err)
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    server := mux.NewServer(":8080", router)
    if err := server.Listen(ctx); err != nil {
        panic(err)
    }
}
```

Run the application:

```bash
go run .
```

Verify the endpoint:

```bash
curl http://localhost:8080/
```

Expected response:

```json
"Mux is working!"
```

`Configure` is the recommended startup path for application setup. It collects route-registration validation errors and returns them explicitly before the server begins serving traffic.

## Development Setup

If you are working on the Mux repository itself:

```bash
git clone https://github.com/fgrzl/mux.git
cd mux
go mod download
go test ./...
```

Useful validation commands while developing:

```bash
go test ./...
go build ./examples/hello-world ./examples/cors-wildcard
```

Nested example modules can be built from their own directories:

```bash
cd examples/todo-api
go build ./...
```

## Troubleshooting

### Go version is too old

Mux follows the Go version declared in `go.mod`. If `go version` reports an older version than `1.25.6`, upgrade Go before building.

### Module download fails

Run:

```bash
go env GOPROXY
go mod tidy
```

If you are behind a corporate proxy, ensure your Go proxy settings are configured correctly.

### Port 8080 is already in use

Use a different port in `mux.NewServer`, for example `:8081`, or stop the process that is already listening on `:8080`.

### TLS files are not found

If you use `mux.WithTLS(...)` or `mux.WithTLSDiscovery(...)`, make sure the certificate and key files exist before startup. `NewServer(...).Listen(ctx)` will fail early if the configured TLS files are invalid.

## Next Steps

- [Quick Start](quick-start.md) for the smallest working API
- [Getting Started](getting-started.md) for a broader introduction
- [Router](router.md) for routing and configuration details
- [WebServer](webserver.md) for production server lifecycle guidance
- [Examples](../examples/) for runnable applications


