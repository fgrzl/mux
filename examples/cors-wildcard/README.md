# CORS Wildcard Example

This example demonstrates the CORS middleware's wildcard pattern support.

## What It Does

The example server:
- ✅ Allows requests from `https://example.com`
- ✅ Allows requests from ANY subdomain of `example.com` (using `*.example.com`)
- ❌ Rejects requests from other domains

## Running the Example

```bash
cd examples/cors-wildcard
go run main.go
```

The server will start on `http://localhost:8080`.

## Testing CORS

### Using curl

Test with an allowed origin:
```bash
# Subdomain will be allowed
curl -H "Origin: https://api.example.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS \
     http://localhost:8080/api/users -i
```

Test with a rejected origin:
```bash
# Evil origin will be rejected (no CORS headers)
curl -H "Origin: https://evil.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS \
     http://localhost:8080/api/users -i
```

### Using a Browser

1. Start the example server: `go run main.go`
2. Open your browser's console on `https://api.example.com` (if you control this domain)
3. Run:
```javascript
fetch('http://localhost:8080/api/users', {
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json'
  }
})
.then(r => r.json())
.then(console.log)
```

## Expected Behavior

### Allowed Origins ✅

These origins will receive CORS headers:
- `https://example.com`
- `https://api.example.com`
- `https://www.example.com`
- `https://staging.example.com`
- `https://v2.api.example.com` (nested subdomains work)
- `http://api.example.com` (different protocol, same domain)
- `https://api.example.com:8443` (with port)

### Rejected Origins ❌

These origins will NOT receive CORS headers:
- `https://evil.com`
- `https://notexample.com` (partial match doesn't count)
- `https://example.org` (different TLD)

## Configuration

The example uses these CORS settings:

```go
mux.UseCORS(router,
    mux.WithAllowedOrigins("https://example.com"),  // Exact domain
    mux.WithOriginWildcard("*.example.com"),        // All subdomains
    mux.WithAllowedMethods("GET", "POST", "PUT", "DELETE"),
    mux.WithAllowedHeaders("Authorization", "Content-Type"),
    mux.WithCredentials(true),
    mux.WithMaxAge(3600),
)
```

## API Endpoints

- `GET /api/users` - Returns a JSON list of users
- `POST /api/users` - Creates a new user

## Real-World Use Cases

This pattern is perfect for:

1. **SaaS Applications**: Allow all customer subdomains
   ```go
   mux.WithOriginWildcard("*.yoursaas.com")
   ```

2. **Multi-Environment Deployments**: Different subdomains for staging, dev, etc.
   ```go
   mux.WithOriginWildcard("*.staging.myapp.com", "*.dev.myapp.com")
   ```

3. **Microservices**: Allow inter-service communication across subdomains
   ```go
   mux.WithOriginWildcard("*.services.internal.com")
   ```

## Security Notes

- The wildcard pattern `*.example.com` does NOT match `example.com` itself (no subdomain)
- Partial matches like `notexample.com` are rejected
- The exact origin is reflected in the response (not the pattern)
- When `WithCredentials(true)`, the origin is always reflected (required for cookies/auth)
