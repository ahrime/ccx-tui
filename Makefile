.PHONY: build run clean test lint

BINARY  := ccx-tui
VERSION := 0.1.0-dev
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY) ./cmd/ccx-tui

run:
	go run ./cmd/ccx-tui

clean:
	rm -rf dist/

test:
	go test ./... -v

lint:
	golangci-lint run ./...
