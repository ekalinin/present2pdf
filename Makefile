.PHONY: build clean test install example

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

