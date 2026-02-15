# present2pdf

A command-line application in Go for converting presentations from `.slide` format (golang.org/x/tools/present) to PDF.

## Features

- ✅ Convert .slide files to PDF
- ✅ Support for headers and text
- ✅ Bulleted lists
- ✅ Code blocks with formatting
- ✅ Author and date information
- ✅ Beautiful slide design
- ✅ Simple command-line interface

## Installation

```bash
# Clone the repository
git clone <repo-url>
cd present2pdf

# Install dependencies and build
make deps
make build

# Or directly via go
go build -o present2pdf ./cmd/present2pdf
```

## Usage

```bash
# Basic usage (output: presentation.pdf)
./present2pdf -input presentation.slide

# With output file specified
./present2pdf -input presentation.slide -output output.pdf

# Use Makefile for example
make example
```

## Command Line Options

- `-input` - path to input .slide file (required)
- `-output` - path to output PDF file (optional, defaults to input filename with .pdf extension)
- `-h` - show help

## .slide File Format

.slide files use a simple text format. Present supports two formats:

- **Markdown-enabled** (recommended): Title starts with `# ` - uses CommonMark syntax
- **Legacy**: Title without `# ` prefix - older syntax

All examples in this project use the **Markdown-enabled** format.

Here's an example structure:

```
# Title of Presentation
Subtitle Goes Here
15 Feb 2026

Author Name
author@example.com

## First Slide Title

This is paragraph text on the slide.

- First bullet point
- Second bullet point
- Third bullet point

## Code Example Slide

Here's some Go code:

	package main

	import "fmt"

	func main() {
		fmt.Println("Hello, World!")
	}

## Another Slide

More content here.
```

### Format Rules

1. **Presentation Header**: The first lines before the first `##` contain metadata:
   - Line 1: `# Title` (the `# ` prefix enables Markdown format)
   - Line 2: Subtitle (optional)
   - Line 3: Date in "DD Mon YYYY" format (e.g., "15 Feb 2026")
   - Following lines: Author information

2. **Slides**: Start with `##` and slide title

3. **Text**: Uses CommonMark Markdown syntax
   - _Italic_: `_text_`
   - **Bold**: `**text**`
   - Inline code: `` `code` ``
   - Links: `[label](url)`

4. **Lists**: Lines starting with `-`

5. **Code**: Blocks with indentation (tabs or spaces)

6. **Comments**: Lines starting with `//` are ignored

For detailed format documentation, see [PRESENT_FORMAT.md](PRESENT_FORMAT.md).

## Examples

The project includes a comprehensive example in the `example/` directory:

- `example/presentation.slide` - full demonstration of all Markdown features

Convert it:

```bash
./present2pdf -input example/presentation.slide
# or
make example
```

## Supported Elements

- ✅ Slide titles
- ✅ Text blocks
- ✅ Bulleted lists
- ✅ Code blocks
- ✅ Author information
- ✅ Dates
- ⚠️ Images (planned)
- ⚠️ Links (planned)
- ⚠️ Videos (planned)

## Project Structure

```
present2pdf/
├── cmd/
│   └── present2pdf/
│       └── main.go            # Application entry point
├── internal/
│   └── converter/
│       ├── converter.go       # Conversion logic
│       └── converter_test.go  # Unit tests
├── example/
│   ├── presentation.slide     # Example presentation
│   └── README.md
├── Makefile                   # Build automation
├── go.mod                     # Go dependencies
└── README.md                  # Documentation
```

## Dependencies

- `golang.org/x/tools/present` - .slide format parsing
- `github.com/jung-kurt/gofpdf` - PDF generation

## Makefile Commands

```bash
make build    # Build the application
make deps     # Install dependencies
make clean    # Remove built files and PDFs
make example  # Build and run on example
make install  # Install to system
make fmt      # Format code
make vet      # Check code
make test     # Run tests
```

## Testing

The project includes comprehensive tests for both Markdown-enabled and legacy formats.

```bash
# Run all tests
make test

# Run with coverage
go test ./... -cover

# Verbose output
go test ./... -v
```

Test coverage: **76.2%** of statements

For detailed testing documentation, see [TESTING.md](TESTING.md).

## License

MIT
