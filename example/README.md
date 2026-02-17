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
# Convert with default light theme
./present2pdf -input example/presentation.slide -output example/presentation.pdf

# Convert with dark theme
./present2pdf -input example/presentation.slide -output example/presentation-dark.pdf -theme dark

# Combine dark theme with nord code style
./present2pdf -input example/presentation.slide -output example/presentation-dark-nord.pdf -theme dark -code-theme nord

# Or use Makefile for default conversion
make example
```

## Theme Examples

You can generate the same presentation with different themes:

### Light Theme (default)
```bash
./present2pdf -input example/presentation.slide -output example/presentation-light.pdf -theme light
```
- Classic light appearance
- White backgrounds
- Blue accents
- Best for printing and bright environments

### Dark Theme
```bash
./present2pdf -input example/presentation.slide -output example/presentation-dark.pdf -theme dark
```
- Modern dark appearance
- Dark backgrounds with light text
- Blue accents optimized for dark mode
- Best for screens and dark environments

### Custom Combinations
```bash
# Dark theme with GitHub code style (light code on dark slides)
./present2pdf -input example/presentation.slide -theme dark -code-theme github

# Light theme with Dracula code style
./present2pdf -input example/presentation.slide -theme light -code-theme dracula
```

For more information about themes, see [PDF_THEMES.md](../docs/PDF_THEMES.md).

## Results

After conversion, PDF files will appear in this directory:
- `presentation.pdf` - default conversion
- `presentation-light.pdf` - light theme example
- `presentation-dark.pdf` - dark theme example
- `presentation-dark-nord.pdf` - dark theme with nord code style

PDF files are ignored by git (see `.gitignore`).

## Format

This presentation uses the **Markdown-enabled present** format with CommonMark syntax.

See [PRESENT_FORMAT.md](../docs/PRESENT_FORMAT.md) for format details.

