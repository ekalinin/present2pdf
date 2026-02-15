# Presentation Example

This directory contains an example `.slide` file to demonstrate the capabilities of `present2pdf`.

## File

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
# Convert the presentation
./present2pdf -input example/presentation.slide -output example/presentation.pdf

# Or use Makefile (converts example by default)
make example
```

## Result

After conversion, the following file will appear in this directory:
- `presentation.pdf` - result of converting presentation.slide

The PDF file is ignored by git (see `.gitignore`).

## Format

This presentation uses the **Markdown-enabled present** format with CommonMark syntax.
See [PRESENT_FORMAT.md](../PRESENT_FORMAT.md) for format details.



