package converter

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/tools/present"
)

//go:embed font/cp1251.map
var cp1251Map []byte

//go:embed font/helvetica_1251.json
var helvetica1251JSON []byte

//go:embed font/helvetica_1251.z
var helvetica1251Z []byte

//go:embed font/jetbrainsmono_1251.json
var jetbrainsmono1251JSON []byte

//go:embed font/jetbrainsmono_1251.z
var jetbrainsmono1251Z []byte

//go:embed font/jetbrainsmono_bold_1251.json
var jetbrainsmono1251BoldJSON []byte

//go:embed font/jetbrainsmono_bold_1251.z
var jetbrainsmono1251BoldZ []byte

// Converter handles conversion from .slide to PDF
type Converter struct {
	pdf                *gofpdf.Fpdf
	translator         func(string) string // UTF-8 translator
	codeTheme          string              // Name of the syntax highlighting style
	theme              Theme               // Color theme for the presentation
	currentSlideTitle  string              // For diagnostic messages
	currentSlideNumber int                 // For diagnostic messages
	quiet              bool                // Suppress diagnostic warnings
}

// Option is a functional option for configuring the Converter
type Option func(*Converter)

// WithCodeTheme sets the code syntax highlighting theme
func WithCodeTheme(themeName string) Option {
	return func(c *Converter) {
		c.codeTheme = themeName
	}
}

// WithTheme sets the PDF color theme
func WithTheme(themeName string) Option {
	return func(c *Converter) {
		if theme, ok := availableThemes[themeName]; ok {
			c.theme = theme
		}
		// If theme not found, keep the default
	}
}

// WithQuiet suppresses diagnostic warnings (slide overflow, code truncation)
func WithQuiet(quiet bool) Option {
	return func(c *Converter) {
		c.quiet = quiet
	}
}

// NewConverter creates a new converter instance with optional configuration
func NewConverter(opts ...Option) *Converter {
	// Default configuration
	c := &Converter{
		codeTheme: "monokai",
		theme:     LightTheme,
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// setTextFont sets the text font with the given style and size
// Uses Helvetica (the only one with proper Cyrillic support). Bold/italic — visual simulation
func (c *Converter) setTextFont(style string, size float64) {
	c.pdf.SetFont("Helvetica", "", size)
}

// setCodeFont sets the code font with the given style and size
func (c *Converter) setCodeFont(style string, size float64) {
	c.pdf.SetFont("JetBrainsMono", style, size)
}

// Convert converts a .slide file to PDF
func (c *Converter) Convert(inputPath, outputPath string) error {
	// Read the slide file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Parse the presentation
	ctx := present.Context{
		ReadFile: func(name string) ([]byte, error) {
			return os.ReadFile(name)
		},
	}

	doc, err := ctx.Parse(strings.NewReader(string(content)), inputPath, 0)
	if err != nil {
		return fmt.Errorf("failed to parse presentation: %w", err)
	}

	// Create temporary directory for font files
	tmpDir, err := os.MkdirTemp("", "present2pdf-fonts-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write embedded font files to temp directory
	fontFiles := map[string][]byte{
		"cp1251.map":                   cp1251Map,
		"helvetica_1251.json":          helvetica1251JSON,
		"helvetica_1251.z":             helvetica1251Z,
		"jetbrainsmono_1251.json":      jetbrainsmono1251JSON,
		"jetbrainsmono_1251.z":         jetbrainsmono1251Z,
		"jetbrainsmono_bold_1251.json": jetbrainsmono1251BoldJSON,
		"jetbrainsmono_bold_1251.z":    jetbrainsmono1251BoldZ,
	}

	for filename, data := range fontFiles {
		if err := os.WriteFile(tmpDir+"/"+filename, data, 0644); err != nil {
			return fmt.Errorf("failed to write font file %s: %w", filename, err)
		}
	}

	// Create PDF with UTF-8 support
	c.pdf = gofpdf.New("L", "mm", "A4", tmpDir)
	c.pdf.SetAutoPageBreak(false, 0)

	// Add Cyrillic fonts with cp1251 encoding
	c.pdf.AddFont("Helvetica", "", "helvetica_1251.json")
	c.pdf.AddFont("Helvetica", "B", "helvetica_1251.json")
	c.pdf.AddFont("Helvetica", "I", "helvetica_1251.json")

	// Add JetBrains Mono for code blocks
	c.pdf.AddFont("JetBrainsMono", "", "jetbrainsmono_1251.json")
	c.pdf.AddFont("JetBrainsMono", "B", "jetbrainsmono_bold_1251.json")

	// Initialize UTF-8 translation for Cyrillic (cp1251)
	tr := c.pdf.UnicodeTranslatorFromDescriptor("cp1251")

	// Store translator for later use
	c.translator = tr

	// Render title slide
	c.currentSlideNumber = 1
	c.renderTitleSlide(doc)

	// Render each section as a slide
	for i, section := range doc.Sections {
		c.currentSlideNumber = i + 2
		c.renderSlide(section)
	}

	// Save PDF
	if err := c.pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}
