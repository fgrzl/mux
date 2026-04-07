# Hello World Example

The smallest useful Mux service: one router, two routes, grouped request accessors, and direct response helpers.

## Run It

```bash
go run .
```

The server listens on `http://localhost:8080`.

## Try It

```bash
curl http://localhost:8080/
```

```json
"Hello, World!"
```

```bash
curl http://localhost:8080/hello/John
```

```json
{
  "message": "Hello, John!",
  "status": "success"
}
```

```bash
curl http://localhost:8080/hello/
```

The last request returns `404` because the route requires a `{name}` segment.

## What It Demonstrates

- `mux.NewRouter()` and `router.Configure(...)`
- Grouped request access with `c.Params().String(...)`, plus `c.OK(...)` and `c.BadRequest(...)`
- `mux.NewServer(...).Listen(ctx)` with signal-driven shutdown

## Next Steps

- [Todo API](../todo-api/) for CRUD, binding, and OpenAPI generation
- [CORS Wildcard](../cors-wildcard/) for middleware configuration
- [WebServer](../webserver/) for health probes and server lifecycle
