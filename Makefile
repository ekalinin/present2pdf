.PHONY: help build clean test install example deps fmt vet
.DEFAULT_GOAL := help

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

# Build application
build:
	go build -o present2pdf ./cmd/present2pdf

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
	go install ./cmd/present2pdf

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Check code
vet:
	go vet ./...

