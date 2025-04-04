# Makefile for a Go project

# Variables
APP_NAME := project0
GO_FILES := $(shell find . -type f -name '*.go')
BUILD_DIR := build
PROTO_FILES:= api/protos
BIN := $(BUILD_DIR)/$(APP_NAME)

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build: $(GO_FILES)
	@echo "Building the application..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BIN) .

# Run the application
.PHONY: run
run: build
	@echo "Running the application..."
	@$(BIN)

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