# Presentation Examples

This directory contains example `.slide` files to demonstrate the capabilities of `present2pdf`.

## Files

### presentation.slide
Comprehensive presentation demonstrating all Markdown features:
- Title slide with author information
- Text formatting (_italic_, **bold**, `code`)
- External links and URLs
- Multiple code examples with comments
- Lists and structured content
- Complex code with error handling
- Practical examples (HTTP server)
- Best practices and resources
- Professional presentation layout

## Usage

From the project root directory, run:

```bash
# Convert the main presentation
./present2pdf -input example/presentation.slide -output example/presentation.pdf

# Or use Makefile
make example
```

## Results

After conversion, the following files will appear in this directory:
- `presentation.pdf` - result of converting presentation.slide
- `syntax_highlight_demo.pdf` - result of converting syntax_highlight_demo.slide

PDF files are ignored by git (see `.gitignore`).

## Format

These presentations use the **Markdown-enabled present** format with CommonMark syntax.

See:
- [PRESENT_FORMAT.md](../PRESENT_FORMAT.md) for format details
- [SYNTAX_HIGHLIGHTING.md](../SYNTAX_HIGHLIGHTING.md) for syntax highlighting details

