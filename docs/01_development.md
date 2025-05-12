# Developer Guide

## Tech stack

Vör is comprised of two main components:

1. CLI client
2. Web server

#### Sqlite DB

Both components persist results to a `sqlite3` database whose path is specified by the `VOER_SQLITEDBPATH` environment
variable. This defaults to `$HOME/.voer/voer.db`

### Web server

There are two processes running on the web server.

#### Frontend

The web UI was created using Templ + DaisyUI + HTMX + Alpine.js.

This is a very flexible framework that enables an extremely rapid development cycle.

#### gRPC API

The gRPC service handles any command send from the CLI client (e.g. validations, package version uploads).

## Dependencies

- Golang

## Linting

- `gofmt`, installed with Go.
- `golangci-lint`, [installed separately](https://golangci-lint.run/welcome/install/#local-installation).

Installation of additional tools:

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
```

### Pre-commit

A utility called [pre-commit](https://pre-commit.com) is used to automatically run our linting utilities before a commit
is made.

```bash
# Install pre-commit hooks
pre-commit install

# Run all linting rules on all files manually
pre-commit run --all-files
```

## Tests

```bash
# Run unit tests
make test
```

## Build

Builds are done with Bazel.

```bash

# Install additional Go CLIs (protoc, templ, etc.)
make install-go-tools

# Build protobufs
make build-proto

# Build CLI
make build

# Or run web server without building
make dev-server
```

## Release

Cut new releases by using tags and GitHub Actions.

```bash

```
