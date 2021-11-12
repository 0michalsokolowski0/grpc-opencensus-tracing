.PHONY: generate-proto
generate-proto:
	protoc --go_out=plugins=grpc:./internal/testpb ./internal/testpb/*.proto

.PHONY: test
test:
	go test -race -coverprofile=coverage.out ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build:
	go build -v ./...
