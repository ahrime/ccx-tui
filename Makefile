.PHONY: build run clean test lint release

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

release:
	@echo "Building all platforms..."
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64   ./cmd/ccx-tui
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64   ./cmd/ccx-tui
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64  ./cmd/ccx-tui
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64  ./cmd/ccx-tui
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe ./cmd/ccx-tui
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-arm64.exe ./cmd/ccx-tui
	@echo "Done. Outputs in dist/:"
	@ls -lh dist/

goreleaser:
	goreleaser release --snapshot --clean
