# Get version info from git
GIT_DESC := $(shell git describe --tags --long --always --dirty 2>/dev/null || echo "unknown")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -I)

# Process version
ifeq ($(findstring -0-g,$(GIT_DESC)),-0-g)
	VERSION_NUM := $(shell echo $(GIT_DESC) | sed 's/-0-g.*//' | sed 's/^v//')
else ifneq ($(findstring -g,$(GIT_DESC)),)
	VERSION_NUM := $(shell echo $(GIT_DESC) | sed -E 's/^v?([^-]*)-([0-9]+)-g.*/\1+\2/')
else
	VERSION_NUM := $(GIT_DESC)
endif

# Complete version string
VERSION_STRING := jdq $(VERSION_NUM) [$(GIT_COMMIT)] ($(BUILD_DATE))

.PHONY: build clean distclean version help fmt lint test unit eas

# Default target
build: jdq

# Build binary only if source files changed
jdq: main.go main_test.go go.mod
	@echo "Building jdq..."
	@echo "$(VERSION_STRING)" > version.txt
	@mkdir -p .go-cache
	@docker run --rm \
		-u "$$(id -u):$$(id -g)" \
		-e HOME=/tmp \
		-e GOCACHE=/tmp/go-cache \
		-e GOMODCACHE=/go/pkg/mod \
		-e CGO_ENABLED=0 \
		-v "$${PWD}":/src \
		-v "$${PWD}/.go-cache":/go/pkg/mod \
		-w /src \
		golang:1.21-alpine sh -c "\
		go mod tidy && \
		go build -ldflags=\"-s -w -extldflags=-static \
			-X 'main.versionString=\$$(cat version.txt)'\" \
		-o jdq main.go"
	@echo "Binary created: ./jdq"

# Show version info
version:
	@echo "$(VERSION_STRING)" > version.txt
	@echo "$(VERSION_STRING)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f jdq version.txt

# Deep clean including test artifacts
distclean: clean
	@echo "Cleaning Go module cache..."
	@chmod -R +w .go-cache 2>/dev/null || true
	@rm -rf .go-cache
	@echo "Running test teardown scripts..."
	@for dir in examples/*/; do \
		if [ -x "$$dir/teardown" ]; then \
			echo "  Running teardown in $$dir"; \
			(cd "$$dir" && ./teardown) || true; \
		fi; \
	done

# Format Go code
fmt:
	@mkdir -p .go-cache
	@docker run --rm \
		-u "$$(id -u):$$(id -g)" \
		-e HOME=/tmp \
		-e GOCACHE=/tmp/go-cache \
		-e GOMODCACHE=/go/pkg/mod \
		-v "$${PWD}":/src \
		-v "$${PWD}/.go-cache":/go/pkg/mod \
		-w /src \
		golang:1.21-alpine go fmt ./...

# Lint Go code
lint:
	@mkdir -p .go-cache
	@docker run --rm \
		-u "$$(id -u):$$(id -g)" \
		-e HOME=/tmp \
		-e GOCACHE=/tmp/go-cache \
		-e GOMODCACHE=/go/pkg/mod \
		-v "$${PWD}":/src \
		-v "$${PWD}/.go-cache":/go/pkg/mod \
		-w /src \
		golangci/golangci-lint:v1.54-alpine golangci-lint run


# Run all tests
test: unit eas

# Run Go unit tests
unit:
	@mkdir -p .go-cache
	@docker run --rm \
		-u "$$(id -u):$$(id -g)" \
		-e HOME=/tmp \
		-e GOCACHE=/tmp/go-cache \
		-e GOMODCACHE=/go/pkg/mod \
		-v "$${PWD}":/src \
		-v "$${PWD}/.go-cache":/go/pkg/mod \
		-w /src \
		golang:1.21-alpine go test -v ./...

# Run Examples as Specifications
eas: jdq
	@./test_examples.sh

# Show help
help:
	@echo "Available targets:"
	@echo "  make build     - Build binary (default)"
	@echo "  make fmt       - Format Go code"
	@echo "  make lint      - Lint Go code"
	@echo "  make test      - Run all tests (unit + eas)"
	@echo "  make unit      - Run Go unit tests"
	@echo "  make eas       - Run Examples as Specifications"
	@echo "  make version   - Show version information"
	@echo "  make clean     - Remove build artifacts"
	@echo "  make distclean - Deep clean including example artifacts"
	@echo "  make help      - Show this help message"
