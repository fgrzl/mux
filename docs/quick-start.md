# Quick Start

Get up and running with Mux in under 5 minutes. This guide shows you the absolute minimum needed to create your first API.

> **New to Mux?** This quick start gets you coding immediately. For a comprehensive tutorial, see the [Interactive Tutorial](interactive-tutorial.md).

## Prerequisites

- Go 1.24.4 or later installed ([Download](https://go.lang.org/dl/))
- Basic familiarity with Go

## Step 1: Create a New Project

```bash
mkdir my-api
cd my-api
go mod init my-api
go get github.com/fgrzl/mux
```

## Step 2: Create Your First API

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
    router := mux.NewRouter().Safe()

    router.GET("/hello", func(c mux.RouteContext) {
        c.OK("Hello, World!")
    })

    if err := router.Err(); err != nil {
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

`Safe()` lets startup configuration accumulate validation errors so `Err()` can fail before the server accepts traffic.

## Step 3: Run and Test

```bash
# Run the server
go run main.go

# In another terminal, test it
curl http://localhost:8080/hello
```

**Output:** `"Hello, World!"`

🎉 **Congratulations!** You have a working API!

## What's Next?

Choose your path:

### Learn by Doing
**→ [Interactive Tutorial](interactive-tutorial.md)** - Build a complete Todo API in 30 minutes with validation, error handling, and OpenAPI documentation.

### Comprehensive Guide
**→ [Getting Started](getting-started.md)** - Step-by-step guide covering all major features with examples.

### Quick Reference
**→ [Cheat Sheet](cheat-sheet.md)** - Copy-paste examples for common patterns.

### Structured Learning
**→ [Learning Path](learning-path.md)** - Progressive 8-level course from beginner to advanced.

## See Also

- [Installation](installation.md) - Detailed setup and requirements
- [Router](router.md) - Routing fundamentals and configuration
- [Middleware](middleware.md) - Built-in middleware guide

Check out the other documentation files to learn about advanced features like authentication, custom middleware, and production deployment patterns.