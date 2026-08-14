.PHONY: build run test lint clean dev docker bench

APP_NAME := gatewayx
BIN_DIR := bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(BUILD_DATE)

build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/gatewayx ./apps/gateway
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/gatewayx-cli ./apps/cli

run: build
	./$(BIN_DIR)/gatewayx

dev:
	go run ./apps/gateway

test:
	go test ./... -v -count=1

test-race:
	go test ./... -race -count=1

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	go vet ./...

clean:
	rm -rf $(BIN_DIR) coverage.out coverage.html

docker:
	docker build -t gatewayx:$(VERSION) .

docker-run:
	docker run -p 8080:8080 -p 9090:9090 \
		-v $(PWD)/gatewayx.example.yaml:/etc/gatewayx/gatewayx.yaml \
		gatewayx:$(VERSION)

install:
	go install ./apps/cli

bench:
	./scripts/bench.sh

deps:
	go mod tidy
	go mod verify
