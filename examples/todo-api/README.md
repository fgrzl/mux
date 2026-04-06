# Todo API Example

A compact CRUD API showing how the public `mux` API fits together in a realistic service.

## Features

- full CRUD operations
- JSON binding and responses
- input validation
- generated OpenAPI 3.1 output
- thread-safe in-memory storage
- `mux.NewServer(...).Listen(ctx)` for startup validation and graceful shutdown

## Run It

```bash
go run .
```

The API listens on `http://localhost:8080`.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/todos` | List todos, optionally filtered with `?completed=true` or `?completed=false` |
| POST | `/todos` | Create a todo |
| GET | `/todos/{id}` | Fetch one todo |
| PUT | `/todos/{id}` | Update a todo |
| DELETE | `/todos/{id}` | Delete a todo |
| GET | `/openapi.json` | Return the generated OpenAPI spec |

## Try It

Create a todo:

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Learn Mux",
    "description": "Complete the interactive tutorial"
  }'
```

List todos:

```bash
curl http://localhost:8080/todos
```

Filter todos:

```bash
curl "http://localhost:8080/todos?completed=true"
```

Fetch one todo:

```bash
curl http://localhost:8080/todos/{id}
```

Update a todo:

```bash
curl -X PUT http://localhost:8080/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'
```

Delete a todo:

```bash
curl -X DELETE http://localhost:8080/todos/{id}
```

## What It Demonstrates

- `router.Configure(...)` as the startup path
- route groups and route-builder metadata
- `c.Bind(...)`, `c.Param(...)`, `c.OK(...)`, and `c.Created(...)`
- OpenAPI generation from registered routes
- server lifecycle with `mux.NewServer(...).Listen(ctx)`

## OpenAPI Output

```bash
curl http://localhost:8080/openapi.json | jq
```

Paste the JSON into [Swagger Editor](https://editor.swagger.io/) if you want an interactive view.

## Next Steps

- [Hello World](../hello-world/) for the smallest possible service
- [Interactive Tutorial](../../docs/interactive-tutorial.md) for a guided build
- [Learning Path](../../docs/learning-path.md) for a broader progression through the framework
