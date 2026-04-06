# Examples

This directory contains runnable example applications that reflect the current public `mux` API.

## Available Examples

| Example | Focus |
|--------|-------|
| [`hello-world/`](hello-world/) | Smallest useful service: routes, params, and direct response helpers |
| [`todo-api/`](todo-api/) | CRUD API with binding, validation, and generated OpenAPI |
| [`cors-wildcard/`](cors-wildcard/) | CORS configuration with exact origins and wildcard subdomains |
| [`redirects/`](redirects/) | Redirect helpers and how to document redirect responses |
| [`openapi-schemas/`](openapi-schemas/) | Struct-tag-driven schema descriptions in generated OpenAPI |
| [`webserver/`](webserver/) | `mux.NewServer(...).Listen(ctx)` with health probes and graceful shutdown |

## Running Examples

Run an example from its directory:

```bash
cd examples/hello-world
go run .
```

Some examples generate files as part of their output. For example, `openapi-schemas/` writes `openapi.json` and `openapi.yaml`.

## Suggested Order

1. [`hello-world/`](hello-world/)
2. [`todo-api/`](todo-api/)
3. [`cors-wildcard/`](cors-wildcard/)
4. [`webserver/`](webserver/)
5. [`redirects/`](redirects/)
6. [`openapi-schemas/`](openapi-schemas/)

## Notes

- Each example has its own README with curl commands or walkthrough notes.
- The test suite compiles every example directory that contains a `main.go`, so the examples are part of the supported public surface.
- For broader guidance, see the main [documentation](../docs/) and the project [README](../README.md).
