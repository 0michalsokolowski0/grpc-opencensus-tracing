.PHONY: generate-test-proto
generate-test-proto:
	protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative ./internal/testpb/test.proto

.PHONY: test
test:
	go test -race -coverprofile=coverage.out ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build:
	go build -v ./...
