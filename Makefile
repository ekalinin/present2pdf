.PHONY: help build clean test install example deps fmt vet release latest-version
.DEFAULT_GOAL := help

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X main.version=$(VERSION)

# Show help information
help:
	@echo "Available commands:"
	@echo "  make help      - Show this help message"
	@echo "  make build     - Build the application"
	@echo "  make deps      - Install and tidy dependencies"
	@echo "  make clean     - Remove built files and generated PDFs"
	@echo "  make example   - Build and run example conversion"
	@echo "  make install   - Install present2pdf to system"
	@echo "  make test      - Run all tests"
	@echo "  make fmt       - Format Go code"
	@echo "  make vet       - Run Go vet for code checks"
	@echo "  make release  - Create tag and push to trigger GitHub release (VERSION=1.0.0)"
	@echo "  make latest-version - Print latest release version from GitHub"
	@echo ""
	@echo "Version control:"
	@echo "  VERSION        - Set version (default: git describe or 'dev')"
	@echo "  Example: make build VERSION=1.0.0"

# Build application
build:
	go build -ldflags "$(LDFLAGS)" -o present2pdf ./cmd/present2pdf

# Install dependencies
deps:
	go mod download
	go mod tidy

# Clean up
clean:
	rm -f present2pdf
	rm -f example/*.pdf

# Test on example
example: build
	./present2pdf -input example/presentation.slide -output example/presentation.pdf
	@echo "PDF created: example/presentation.pdf"

# Install to system
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/present2pdf

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Check code
vet:
	go vet ./...

# Create tag and push to trigger GitHub Actions release
# Usage: make release VERSION=1.0.0
release:
	@v="$(VERSION)" && \
	if [ -z "$$v" ] || [ "$$v" = "dev" ] || echo "$$v" | grep -qE 'dirty|-g[0-9a-f]'; then \
		echo "Error: VERSION must be explicit semver. Example: make release VERSION=1.0.0"; exit 1; \
	fi && \
	tag="v$$(echo $(VERSION) | sed 's/^v//')" && \
	echo "Creating tag $$tag and pushing to trigger GitHub release..." && \
	git tag -a $$tag -m "Release $$tag" && \
	git push origin $$tag

# Get latest release version from GitHub
latest-version:
	@v=$$(curl -sL "https://api.github.com/repos/ekalinin/present2pdf/releases/latest" | \
		grep -o '"tag_name": *"[^"]*"' | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/'); \
	if [ -n "$$v" ]; then echo $$v; else echo "No releases found"; fi