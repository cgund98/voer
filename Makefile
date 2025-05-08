
install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

build-proto:
	protoc -I. --go_out=. --go-grpc_out=. api/v1/package.proto


build:
	# Build protobuf
	make build-proto

	# Build voer
	mkdir -p dist
	go build -o dist/voer cmd/voer/main.go
