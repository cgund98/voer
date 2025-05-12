.PHONY: all test clean

install-go-tools:
	# Install protobuf tools
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

	# Templ
	go install github.com/a-h/templ/cmd/templ@latest

build-proto:
	protoc -I. --go_out=. --go-grpc_out=. api/v1/package.proto


build-cli:
	# Build protobuf
	make build-proto

	# Build templ
	templ generate

	# Build voer
	mkdir -p dist
	go build -o dist/voer cmd/voer/main.go

test:
	go test -v ./...


dev-server:
	templ generate
	go run cmd/voer/main.go server --grpc-port 8000 --frontend-port 8080
