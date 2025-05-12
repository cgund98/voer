## Go builder
FROM golang:1.23.5-bookworm AS builder

ENV DEBIAN_FRONTEND=noninteractive
ENV PROTOC_GEN_GO_VERSION 1.34.1
ENV PROTOC_GEN_GO_GRPC_VERSION 1.4.0

# Install protoc
RUN apt-get update && apt-get install -y protobuf-compiler && apt-get clean && rm -rf /var/lib/apt/lists/*

# Install protoc-gen-go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v${PROTOC_GEN_GO_VERSION}

# Install protoc-gen-go-grpc
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v${PROTOC_GEN_GO_GRPC_VERSION}

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY api ./api
COPY cmd ./cmd
COPY internal ./internal

RUN templ generate
RUN protoc -I. --go_out=. --go-grpc_out=. api/v1/package.proto

WORKDIR /app/cmd/voer
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/voer

## Production image
FROM debian:bookworm-slim AS prod

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y \
    ca-certificates curl traceroute && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/voer /app/voer

ENTRYPOINT ["/app/voer", "server"]
