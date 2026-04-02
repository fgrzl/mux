# Installation Guide

This guide covers the installation and setup requirements for the Mux HTTP router library.

## Requirements

### Go Version
Mux requires **Go 1.24.4** or later. Check your Go version:

```bash
go version
```

If you need to upgrade Go, visit the [official Go downloads page](https://golang.org/dl/).

### System Requirements
- **Operating System**: Linux, macOS, or Windows
- **Memory**: Minimum 512MB RAM available
- **Network**: Internet connection for downloading dependencies

## Installation

### Using Go Modules (Recommended)

Add Mux to your project using Go modules:

```bash
go get github.com/fgrzl/mux
```

### Initialize a New Project

Create a new Go project with Mux:

```bash
mkdir my-api
cd my-api
go mod init my-api
go get github.com/fgrzl/mux
```

## Dependencies

Mux automatically manages its dependencies through Go modules. Key dependencies include:

### Core Dependencies
- `github.com/fgrzl/claims` - JWT claims handling
- `github.com/fgrzl/json` - JSON processing
- `github.com/google/uuid` - UUID generation and parsing
- `gopkg.in/yaml.v3` - YAML processing for OpenAPI specs

### Optional Dependencies
These are included but only used when specific features are enabled:

- `github.com/oschwald/geoip2-golang` - GeoIP support for export control
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` - OpenTelemetry tracing
- `github.com/stretchr/testify` - Testing framework (development only)

## Verification

Verify your installation by creating a simple "Hello World" application:

### Create main.go
```go
package main

import (
    "net/http"
    "github.com/fgrzl/mux"
)

func main() {
    router := mux.NewRouter()
    
    router.GET("/", func(c mux.RouteContext) {
        c.OK("Mux is working!")
    })
    
    http.ListenAndServe(":8080", router)
}
```

### Run the Application
```bash
go run main.go
```

### Test the Endpoint
```bash
curl http://localhost:8080/
```

You should see the response: `"Mux is working!"`

## Development Setup

For development and contributing to Mux:

### Clone the Repository
```bash
git clone https://github.com/fgrzl/mux.git
cd mux
```

### Install Development Dependencies
```bash
go mod download
```

### Run Tests
```bash
go test ./... -v
```

### Run with Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## IDE Setup

### VS Code
Install the Go extension for VS Code:
1. Open VS Code
2. Go to Extensions (Ctrl+Shift+X)
3. Search for "Go" and install the official Go extension
4. Run `Go: Install/Update Tools` from the command palette

### GoLand/IntelliJ IDEA
GoLand has built-in Go support. For IntelliJ IDEA:
1. Install the Go plugin
2. Configure Go SDK in Project Settings
3. Enable Go modules integration

## Docker Setup (Optional)

Create a Dockerfile for containerized deployment:

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
```

### Build and Run
```bash
docker build -t my-api .
docker run -p 8080:8080 my-api
```

## Troubleshooting

### Common Issues

#### "cannot find module" Error
```bash
# Ensure you're in the correct directory
go mod init your-project-name
go get github.com/fgrzl/mux
```

#### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080
# Kill the process
kill -9 <PID>
```

#### Go Version Too Old
```bash
# Check minimum version requirement
go version  # Should be 1.24.4+
```

#### Permission Denied (Linux/macOS)
```bash
# Run with different port (> 1024)
http.ListenAndServe(":8080", router)  # Instead of :80
```

### Getting Help

- **Documentation**: Check the [docs/](../docs/) directory
- **Issues**: Report bugs on [GitHub Issues](https://github.com/fgrzl/mux/issues)
- **Examples**: See working examples in the [examples/](../examples/) directory

## Next Steps

Once Mux is installed:

1. Read the [Quick Start Tutorial](quick-start.md)
2. Explore the [Router Documentation](router.md)
3. Check out [Example Applications](../examples/)
4. Review the [API Reference](api-reference.md)

## Environment Variables

Mux doesn't require environment variables for basic operation, but some features can be configured:

```bash
# Optional: Set log level for development
export LOG_LEVEL=debug

# Optional: Set GeoIP database path for export control
export GEOIP_DB_PATH=/path/to/GeoLite2-Country.mmdb

# Optional: OpenTelemetry configuration
export OTEL_SERVICE_NAME=my-api
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

## Performance Tuning

For production deployments, consider these Go runtime settings:

```bash
# Set maximum number of OS threads
export GOMAXPROCS=4

# Enable memory ballast for stable GC performance
export GOGC=100

# Disable CGO for static binaries (if not using GeoIP)
CGO_ENABLED=0 go build
```

## See Also

- [Quick Start](quick-start.md) - Get running in 5 minutes
- [Getting Started](getting-started.md) - Comprehensive introduction
- [Router](router.md) - Routing fundamentals
- [WebServer](webserver.md) - Production server setup
- [Best Practices](best-practices.md) - Deployment patterns