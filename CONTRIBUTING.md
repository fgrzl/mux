# Contributing

Thanks for contributing to mux.

## Setup

1. Fork and clone the repository.
2. `go mod download`
3. `go test ./...`

## Pull requests

- Run `go fmt ./...` and `go vet ./...`.
- Update `docs/` and examples under `examples/` for user-facing changes.
- Follow guidelines in `docs/dev/` when adding tests or benchmarks.
- Keep OpenAPI generation behavior backward compatible in minor releases.

## Changelog

Note changes under `## [Unreleased]` in [CHANGELOG.md](CHANGELOG.md).
