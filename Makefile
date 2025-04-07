# Makefile for a Go project

# Variables
APP_NAME := project0
GO_FILES := $(shell find . -type f -name '*.go')
BUILD_DIR := cmd/sber_loan
PROTO_FILES:= api/protos
BIN := $(BUILD_DIR)/$(APP_NAME)

# Default target
.PHONY: all
all: build

# Build the application container
.PHONY: build
build: 
	docker build -t sber-grpc-server .

# Run the application container
.PHONY: run
run: 
	docker run -p 8080:8080 -d -v ./config/config.yml:/app/config/config.yml sber-grpc-server 
# Stop the application container
.PHONY: stop
stop: 
	docker stop sber-grpc-server

# Delete the application container
.PHONY: rm
rm: 
	docker rm sber-grpc-server


# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
.PHONY: lint
lint:
	golangci-lint cache clean
	golangci-lint run --enable=gofmt --fix --config .golangci.yaml

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod tidy

 #generate on proto
.PHONY: proto
proto:
	protoc -I. -Ithird_party \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=./ \
		$(PROTO_FILES)/**/*.proto