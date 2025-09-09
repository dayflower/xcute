.PHONY: build test clean install snapshot release help

# Build the binary for current platform
build:
	go build -o xcute

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Clean build artifacts
clean:
	rm -f xcute
	rm -rf dist/

# Install binary to $GOPATH/bin
install:
	go install

# Build snapshot with goreleaser (all platforms)
snapshot:
	goreleaser build --snapshot --clean

# Release with goreleaser (requires git tag)
release:
	goreleaser release --clean

# Format code
fmt:
	go fmt ./...

# Run linter (if installed)
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Run all checks (test + fmt + tidy)
check: test fmt tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  build        Build binary for current platform"
	@echo "  test         Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  clean        Clean build artifacts"
	@echo "  install      Install binary to \$$GOPATH/bin"
	@echo "  snapshot     Build snapshot for all platforms"
	@echo "  release      Release with goreleaser"
	@echo "  fmt          Format code"
	@echo "  lint         Run linter"
	@echo "  tidy         Tidy dependencies"
	@echo "  check        Run all checks (test + fmt + tidy)"
	@echo "  help         Show this help"