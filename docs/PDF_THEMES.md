# PDF Themes

present2pdf supports different color themes for PDF presentations, allowing you to customize the overall appearance of the document.

## Available Themes

### Light (default theme)

Classic light theme with contrasting elements:

- **Title slide**: blue background (#2980B9)
- **Title slide headings**: white text
- **Content slides**: white background
- **Slide titles**: blue text (#2980B9)
- **Body text**: black
- **Code blocks**: dark background (#282C34)

**Usage:**

```bash
./present2pdf -input presentation.slide -theme light
# or simply (light is the default theme)
./present2pdf -input presentation.slide
```

### Dark (dark theme)

Modern dark theme for comfortable viewing in low-light conditions:

- **Title slide**: dark blue background (#1E1E2E)
- **Title slide headings**: light gray text (#CDD6F4)
- **Content slides**: dark gray background (#24273A)
- **Slide titles**: light blue text (#89B4FA)
- **Body text**: light gray (#CDD6F4)
- **Code blocks**: very dark background (#1E1E2E)

**Usage:**

```bash
./present2pdf -input presentation.slide -theme dark
```

## Combining Themes

You can combine PDF themes with code highlighting themes:

```bash
# Dark PDF with GitHub code style
./present2pdf -input presentation.slide -theme dark -code-theme github

# Light PDF with Dracula code style
./present2pdf -input presentation.slide -theme light -code-theme dracula

# Dark PDF with Nord code style
./present2pdf -input presentation.slide -theme dark -code-theme nord
```

## List of Available Themes

To see all available PDF themes:

```bash
./present2pdf -list-themes
```

To see all available code highlighting themes:

```bash
./present2pdf -list-code-themes
```

## Theme Structure

Each theme defines the following colors:

```go
type Theme struct {
    // Title slide colors
    TitleBackground RGB
    TitleText       RGB
    TitleSubtext    RGB
    TitleDate       RGB

    // Content slide colors
    SlideBackground RGB
    SlideTitle      RGB
    SlideTitleLine  RGB
    SlideText       RGB

    // Code block colors
    CodeBackground  RGB
    CodeText        RGB
    CodeLineNumber  RGB
}
```

## Creating Custom Themes

To add your own theme, edit the file `internal/converter/converter.go`:

1. Define a new theme as a variable of type `Theme`
2. Add it to the `availableThemes` map
3. Rebuild the application

Example:

```go
CustomTheme = Theme{
    TitleBackground: RGB{20, 20, 30},
    TitleText:       RGB{255, 255, 255},
    // ... other colors
}

availableThemes = map[string]Theme{
    "light":  LightTheme,
    "dark":   DarkTheme,
    "custom": CustomTheme,
}
```

## Using the API Directly

If you're using present2pdf as a library, you can use the functional options pattern:

```go
import "github.com/ekalinin/present2pdf/internal/converter"

// Default configuration (light theme, monokai code style)
conv := converter.NewConverter()

// With custom code theme
conv := converter.NewConverter(
    converter.WithCodeTheme("github"),
)

// With custom PDF theme
conv := converter.NewConverter(
    converter.WithTheme("dark"),
)

// With both options
conv := converter.NewConverter(
    converter.WithCodeTheme("dracula"),
    converter.WithTheme("dark"),
)

// Convert
err := conv.Convert("presentation.slide", "output.pdf")
```

## Recommendations

- **Light theme** is suitable for:
  - Printing on paper
  - Presentations in well-lit rooms
  - Projectors with low brightness

- **Dark theme** is suitable for:
  - Viewing on screens in dark rooms
  - Evening presentations
  - Modern LED screens with high contrast
  - Energy saving on OLED screens
