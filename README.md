# present2pdf

A command-line application in Go for converting presentations from `.slide` format (golang.org/x/tools/present) to PDF.

## Features

- ✅ Convert .slide files to PDF
- ✅ Support for headers and text
- ✅ Bulleted lists
- ✅ Code blocks with syntax highlighting
- ✅ Full UTF-8 support
- ✅ Customizable code highlighting themes (80+ themes available)
- ✅ PDF color themes (light and dark)
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

# With custom code highlighting theme
./present2pdf -input presentation.slide -code-theme dracula

# With custom PDF theme
./present2pdf -input presentation.slide -theme dark

# Combine code and PDF themes
./present2pdf -input presentation.slide -code-theme github -theme dark

# List available code highlighting themes
./present2pdf -list-code-themes

# List available PDF themes
./present2pdf -list-themes

# Use Makefile for example
make example
```

## Command Line Options

- `-input` - path to input .slide file (required)
- `-output` - path to output PDF file (optional, defaults to input filename with .pdf extension)
- `-code-theme` - code syntax highlighting theme (optional, default: `monokai`)
- `-theme` - PDF color theme: `light` or `dark` (optional, default: `light`)
- `-list-code-themes` - list all available code highlighting themes and exit
- `-list-themes` - list all available PDF themes and exit
- `-version` - show version information and exit
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

For detailed format documentation, see [PRESENT_FORMAT.md](docs/PRESENT_FORMAT.md).

## Examples

The project includes comprehensive examples in the `example/` directory:

- `example/presentation.slide` - full demonstration of all Markdown features
- `example/cyrillic_demo.slide` - full demonstration of Cyrillic support

Convert them:

```bash
# Convert main example
./present2pdf -input example/presentation.slide
# or
make example

# Convert Cyrillic example
./present2pdf -input example/cyrillic_demo.slide -output example/cyrillic_demo.pdf
./present2pdf -input example/cyrillic_demo.slide -output example/cyrillic_demo_dark.pdf -theme dark
```

## UTF-8 Support

The tool has **full UTF-8 and Cyrillic support** (Russian, Ukrainian, Belarusian, etc.) in all elements:

- ✅ Titles and subtitles in Cyrillic
- ✅ Slide content in Cyrillic with **bold** and _italic_ formatting
- ✅ Lists with Cyrillic text and formatting
- ✅ **Code blocks with Cyrillic support** (comments, strings, variable names)

See `example/cyrillic_demo.slide` for a complete demonstration.

**Fonts used:**
- **Helvetica** (Arial) - main text font with full Cyrillic support; bold/italic simulated visually
- **JetBrains Mono** - monospace font for code blocks with excellent Cyrillic support

For details about code fonts see [MONOSPACE_FONT.md](docs/MONOSPACE_FONT.md).

## Supported Elements

- ✅ Slide titles
- ✅ Text blocks with **bold** and _italic_ formatting
- ✅ Bulleted lists with formatting
- ✅ Code blocks with syntax highlighting
- ✅ Author information
- ✅ Dates
- ⚠️ Images (planned)
- ⚠️ Links (planned)
- ⚠️ Videos (planned)

### Syntax Highlighting

Code blocks in your slides are automatically highlighted with proper syntax colors. The tool supports:

- **Automatic language detection** from file extensions in code blocks
- **Multiple languages**: Go, Python, JavaScript, TypeScript, Java, C, C++, Rust, Ruby, PHP, Bash, HTML, CSS, JSON, XML, YAML, SQL, and more
- **Customizable color schemes**: Choose from 70+ themes including monokai, github, dracula, vim, solarized, and more
- **Fallback to plain rendering** if highlighting fails

#### Available Code Themes

Run `./present2pdf -list-code-themes` to see all available options. Popular themes include:

- `monokai` (default) - Dark theme with vibrant colors
- `github` - Light theme matching GitHub's style
- `dracula` - Dark theme with purple and cyan accents
- `solarized-dark` / `solarized-light` - Popular color schemes
- `vim` - Classic Vim colors
- `nord` - Arctic-inspired color palette
- And 60+ more!

### PDF Themes

The tool supports different color themes for the PDF presentation itself, allowing you to customize the overall look and feel:

#### Light Theme (default)

- Blue title slide background
- White slide backgrounds
- Blue titles and accents
- Black text
- Dark code blocks with syntax highlighting

#### Dark Theme

- Dark blue-gray backgrounds
- Light text on dark backgrounds
- Blue accents
- Optimized for dark environments
- Matches modern dark mode aesthetics

**Usage:**

```bash
# Use light theme (default)
./present2pdf -input presentation.slide -theme light

# Use dark theme
./present2pdf -input presentation.slide -theme dark

# List all available PDF themes
./present2pdf -list-themes
```

You can combine PDF themes with code highlighting themes:

```bash
# Dark PDF with GitHub code style
./present2pdf -input presentation.slide -theme dark -code-theme github
```

For more details, see [PDF_THEMES.md](docs/PDF_THEMES.md).

#### Usage

```bash
# Use default monokai theme
./present2pdf -input presentation.slide

# Use GitHub theme (light background)
./present2pdf -input presentation.slide -code-theme github

# Use Dracula theme
./present2pdf -input presentation.slide -code-theme dracula -output output.pdf
```

Example code block:

```
## Code Example

	package main

	import "fmt"

	func main() {
		fmt.Println("Hello, World!")
	}
```

The code will be rendered with syntax highlighting in the PDF output.

## Project Structure

```
present2pdf/
├── cmd/
│   └── present2pdf/
│       └── main.go            # Application entry point
├── docs/                      # Additional documentation
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
- `github.com/alecthomas/chroma/v2` - Syntax highlighting

## Makefile Commands

```bash
make build          # Build the application
make deps           # Install dependencies
make clean          # Remove built files and PDFs
make example        # Build and run on example
make install        # Install to system
make fmt            # Format code
make vet            # Check code
make test           # Run tests
```

### Building with Version

By default, the version is determined automatically from git tags:

```bash
# Build with auto-detected version (from git describe)
make build

# Build with specific version
make build VERSION=1.0.0

# Install with specific version
make install VERSION=1.0.0
```

To check the version:

```bash
./present2pdf -version
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

## License

[MIT](LICENSE)
