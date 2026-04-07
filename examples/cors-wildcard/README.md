# CORS Wildcard Example

This example shows how to allow an exact origin and any matching subdomain with the built-in CORS middleware.

## What It Does

The server:
- allows requests from `https://example.com`
- allows requests from any subdomain that matches `*.example.com`
- rejects origins outside that allowlist

## Run It

```bash
go run .
```

The server listens on `http://localhost:8080` and shuts down cleanly on `Ctrl+C`.

## Test CORS

### Allowed origin

```bash
curl -H "Origin: https://api.example.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS \
     http://localhost:8080/api/users -i
```

### Rejected origin

```bash
curl -H "Origin: https://evil.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS \
     http://localhost:8080/api/users -i
```

## Expected Behavior

Allowed:
- `https://example.com`
- `https://api.example.com`
- `https://www.example.com`
- `https://staging.example.com`
- `https://v2.api.example.com`
- `http://api.example.com`
- `https://api.example.com:8443`

Rejected:
- `https://evil.com`
- `https://notexample.com`
- `https://example.org`

## Configuration

```go
mux.UseCORS(router,
    mux.WithCORSAllowedOrigins("https://example.com"),
    mux.WithCORSOriginWildcard("*.example.com"),
    mux.WithCORSAllowedMethods("GET", "POST", "PUT", "DELETE"),
    mux.WithCORSAllowedHeaders("Authorization", "Content-Type"),
    mux.WithCORSCredentials(true),
    mux.WithCORSMaxAge(3600),
)
```

## API Endpoints

- `GET /api/users`
- `POST /api/users`

## Security Notes

- `*.example.com` does not match `example.com`; include the root origin explicitly when you need both.
- Partial matches like `notexample.com` are rejected.
- When `WithCORSCredentials(true)` is enabled, the middleware reflects the matched origin instead of returning `*`.
