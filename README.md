# Vör

Schema Registry for Protobufs

### Why?

### How it works

1. Developer creates a new `.proto` package.
2. When ready, developer publishes package to registry.
3. Consumers pull package schema from registry.

Over time, the schema owner can publish new package versions, while the Schema Registry will ensure that no breaking
changes are made.

### Example use case

## Usage

## Development

### Dependencies

### Linting

- `gofmt`, installed with Go.
- `golangci-lint`, [installed separately](https://golangci-lint.run/welcome/install/#local-installation).

Installation of additional tools:

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
```

#### Pre-commit

A utility called [pre-commit](https://pre-commit.com) is used to automatically run our linting utilities before a commit
is made.

```bash
# Install pre-commit hooks
pre-commit install

# Run all linting rules on all files manually
pre-commit run --all-files
```

### Tests

```bash
# Run unit tests
make test
```

### Build

Builds are done with Bazel.

```bash

# Install additional Go CLIs (protoc, templ, etc.)
make install-go-tools

# Build protobufs
make build-proto

# Build CLI
make build
```

### Release

Cut new releases by using tags and GitHub Actions.

```bash

```
