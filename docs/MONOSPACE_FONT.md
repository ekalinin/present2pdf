# Monospace Font with Cyrillic Support

## Current Status

The project uses **JetBrains Mono** - a true monospace font with excellent Cyrillic support (cp1251 encoding). The font is embedded in the project and works out of the box.

## Built-in Fonts

The project includes the following built-in fonts:

### JetBrains Mono (used for code)
- ✅ Monospace font specifically designed for programming
- ✅ Excellent Cyrillic support
- ✅ Modern design, great readability
- ✅ Embedded in the project (Regular and Bold)
- 📦 Version: 2.304

### Helvetica (used for text)
- ✅ Proportional font for main text
- ✅ Cyrillic support via cp1251
- ✅ Used for headings, text, and lists

## How to Add Another Monospace Font

If you need a different monospace font for code with Cyrillic support, follow these steps:

### 1. Choose a TTF Font

Recommended monospace fonts with Cyrillic support:
- **DejaVu Sans Mono** - free, excellent Cyrillic support
- **JetBrains Mono** - specifically for code, free
- **Fira Code** - with ligatures, free
- **Consolas** - from Microsoft
- **Liberation Mono** - free Courier alternative

### 2. Convert TTF to gofpdf Format

Use the `makefont` utility from the gofpdf package:

```bash
go get github.com/jung-kurt/gofpdf/makefont
makefont --embed --enc=cp1251 DejaVuSansMono.ttf DejaVuSansMono-Bold.ttf
```

This will create files:
- `DejaVuSansMono.json` - font metrics
- `DejaVuSansMono.z` - compressed font

### 3. Add Files to Project

Copy the created files to `internal/converter/font/`:
```
internal/converter/font/
├── cp1251.map
├── helvetica_1251.json
├── helvetica_1251.z
├── dejavusansmono_1251.json  <- new
└── dejavusansmono_1251.z     <- new
```

### 4. Update Code

In `internal/converter/converter.go` add:

```go
//go:embed font/dejavusansmono_1251.json
var dejavusansmono1251JSON []byte

//go:embed font/dejavusansmono_1251.z
var dejavusansmono1251Z []byte
```

Add to `fontFiles`:
```go
fontFiles := map[string][]byte{
    "cp1251.map":                 cp1251Map,
    "helvetica_1251.json":        helvetica1251JSON,
    "helvetica_1251.z":           helvetica1251Z,
    "dejavusansmono_1251.json":   dejavusansmono1251JSON,
    "dejavusansmono_1251.z":      dejavusansmono1251Z,
}
```

Register the font:
```go
c.pdf.AddFont("DejaVuSansMono", "", "dejavusansmono_1251.json")
c.pdf.AddFont("DejaVuSansMono", "B", "dejavusansmono_1251.json")
```

Replace in code rendering functions:
```go
// Before:
c.pdf.SetFont("Helvetica", "", 9)

// After:
c.pdf.SetFont("DejaVuSansMono", "", 9)
```

### 5. Rebuild the Project

```bash
go build ./...
```

## Note

Adding a TTF font will increase the binary size by approximately 100-300 KB (compressed).

## Alternative

If binary size is critical, you can keep Helvetica. It reads well and supports Cyrillic, although it is not monospace.
