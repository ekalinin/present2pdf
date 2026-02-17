package converter

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
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

// RGB represents an RGB color
type RGB struct {
	R, G, B int
}

// Theme defines the color scheme for PDF presentation
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
	CodeBackground RGB
	CodeText       RGB
	CodeLineNumber RGB

	// Link color
	LinkColor RGB
}

// Predefined themes
var (
	// LightTheme is the default light theme
	LightTheme = Theme{
		TitleBackground: RGB{41, 128, 185},  // Blue
		TitleText:       RGB{255, 255, 255}, // White
		TitleSubtext:    RGB{255, 255, 255}, // White
		TitleDate:       RGB{255, 255, 255}, // White
		SlideBackground: RGB{255, 255, 255}, // White
		SlideTitle:      RGB{41, 128, 185},  // Blue
		SlideTitleLine:  RGB{41, 128, 185},  // Blue
		SlideText:       RGB{0, 0, 0},       // Black
		CodeBackground:  RGB{40, 44, 52},    // Dark gray
		CodeText:        RGB{171, 178, 191}, // Light gray
		CodeLineNumber:  RGB{128, 128, 128}, // Gray
		LinkColor:       RGB{0, 102, 204},   // Link blue
	}

	// DarkTheme is a dark theme
	DarkTheme = Theme{
		TitleBackground: RGB{30, 30, 46},    // Dark blue-gray
		TitleText:       RGB{205, 214, 244}, // Light gray
		TitleSubtext:    RGB{166, 173, 200}, // Medium gray
		TitleDate:       RGB{137, 180, 250}, // Light blue
		SlideBackground: RGB{36, 39, 58},    // Dark gray-blue
		SlideTitle:      RGB{137, 180, 250}, // Light blue
		SlideTitleLine:  RGB{137, 180, 250}, // Light blue
		SlideText:       RGB{205, 214, 244}, // Light gray
		CodeBackground:  RGB{30, 30, 46},    // Darker blue-gray
		CodeText:        RGB{205, 214, 244}, // Light gray
		CodeLineNumber:  RGB{108, 112, 134}, // Medium gray
		LinkColor:       RGB{137, 180, 250}, // Light blue
	}

	// availableThemes maps theme names to themes
	availableThemes = map[string]Theme{
		"light": LightTheme,
		"dark":  DarkTheme,
	}
)

// Converter handles conversion from .slide to PDF
type Converter struct {
	pdf               *gofpdf.Fpdf
	translator        func(string) string // UTF-8 translator
	codeTheme         string              // Name of the syntax highlighting style
	theme             Theme               // Color theme for the presentation
	currentSlideTitle  string              // For diagnostic messages
	currentSlideNumber int                 // For diagnostic messages
	quiet              bool                // Suppress diagnostic warnings
}

// Token represents a syntax-highlighted token
type Token struct {
	Type  chroma.TokenType
	Value string
	Color [3]int // RGB color
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

// GetAvailableStyles returns a list of available syntax highlighting styles
func GetAvailableStyles() []string {
	return styles.Names()
}

// GetAvailableThemes returns a list of available PDF themes
func GetAvailableThemes() []string {
	themes := make([]string, 0, len(availableThemes))
	for name := range availableThemes {
		themes = append(themes, name)
	}
	return themes
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
		"cp1251.map":                       cp1251Map,
		"helvetica_1251.json":              helvetica1251JSON,
		"helvetica_1251.z":                 helvetica1251Z,
		"jetbrainsmono_1251.json":          jetbrainsmono1251JSON,
		"jetbrainsmono_1251.z":             jetbrainsmono1251Z,
		"jetbrainsmono_bold_1251.json":     jetbrainsmono1251BoldJSON,
		"jetbrainsmono_bold_1251.z":        jetbrainsmono1251BoldZ,
	}

	for filename, content := range fontFiles {
		if err := os.WriteFile(tmpDir+"/"+filename, content, 0644); err != nil {
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

// renderTitleSlide renders the title page
func (c *Converter) renderTitleSlide(doc *present.Doc) {
	c.pdf.AddPage()

	// Background
	c.pdf.SetFillColor(c.theme.TitleBackground.R, c.theme.TitleBackground.G, c.theme.TitleBackground.B)
	c.pdf.Rect(0, 0, 297, 210, "F")

	// Title
	c.pdf.SetTextColor(c.theme.TitleText.R, c.theme.TitleText.G, c.theme.TitleText.B)
	c.setTextFont("B", 54)
	c.pdf.SetXY(20, 70)
	c.pdf.MultiCell(257, 23, c.translator(doc.Title), "", "C", false)

	// Subtitle
	if doc.Subtitle != "" {
		c.pdf.SetTextColor(c.theme.TitleSubtext.R, c.theme.TitleSubtext.G, c.theme.TitleSubtext.B)
		c.setTextFont("", 30)
		c.pdf.SetXY(20, 95)
		c.pdf.MultiCell(257, 15, c.translator(doc.Subtitle), "", "C", false)
	}

	// Authors
	if len(doc.Authors) > 0 {
		c.pdf.SetTextColor(c.theme.TitleSubtext.R, c.theme.TitleSubtext.G, c.theme.TitleSubtext.B)
		c.setTextFont("", 21)
		y := 130.0
		for _, author := range doc.Authors {
			authorText := c.extractAuthorText(author)
			if authorText != "" {
				c.pdf.SetXY(20, y)
				c.pdf.MultiCell(257, 12, c.translator(authorText), "", "C", false)
				y += 15
			}
		}
	}

	// Date
	if !doc.Time.IsZero() {
		c.pdf.SetTextColor(c.theme.TitleDate.R, c.theme.TitleDate.G, c.theme.TitleDate.B)
		c.setTextFont("I", 18)
		c.pdf.SetXY(20, 180)
		c.pdf.MultiCell(257, 9, c.translator(doc.Time.Format("January 2, 2006")), "", "C", false)
	}
}

// renderSlide renders a single slide
func (c *Converter) renderSlide(section present.Section) {
	c.currentSlideTitle = section.Title
	c.pdf.AddPage()

	// Background
	c.pdf.SetFillColor(c.theme.SlideBackground.R, c.theme.SlideBackground.G, c.theme.SlideBackground.B)
	c.pdf.Rect(0, 0, 297, 210, "F")

	// Title
	c.pdf.SetTextColor(c.theme.SlideTitle.R, c.theme.SlideTitle.G, c.theme.SlideTitle.B)
	c.setTextFont("B", 29)
	c.pdf.SetXY(20, 15)
	c.pdf.MultiCell(257, 12, c.translator(section.Title), "", "L", false)

	// Draw a line under the title
	c.pdf.SetDrawColor(c.theme.SlideTitleLine.R, c.theme.SlideTitleLine.G, c.theme.SlideTitleLine.B)
	c.pdf.SetLineWidth(0.5)
	c.pdf.Line(20, 36, 277, 36)

	// Content
	c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
	y := 45.0

	for _, elem := range section.Elem {
		y = c.renderElement(elem, y)
		if y > 190 {
			if !c.quiet {
				fmt.Fprintf(os.Stderr, "Warning: slide %d \"%s\" does not fit - content overflow (y=%.0f), some elements cut off\n", c.currentSlideNumber, section.Title, y)
			}
			break // Avoid content overflow
		}
	}
}

// renderElement renders a single element
func (c *Converter) renderElement(elem present.Elem, y float64) float64 {
	switch e := elem.(type) {
	case present.Text:
		return c.renderText(e, y)
	case present.List:
		return c.renderList(e, y)
	case present.Code:
		return c.renderCode(e, y)
	case present.HTML:
		return c.renderHTML(e, y)
	case present.Link:
		return c.renderLink(e, y)
	default:
		// Skip unsupported elements
		return y
	}
}

// renderLink renders a .link directive as a clickable hyperlink
func (c *Converter) renderLink(link present.Link, y float64) float64 {
	label := link.Label
	urlStr := ""
	if link.URL != nil {
		urlStr = link.URL.String()
	}
	if label == "" {
		label = urlStr
	}

	c.setTextFont("", 18)
	c.pdf.SetTextColor(c.theme.LinkColor.R, c.theme.LinkColor.G, c.theme.LinkColor.B)

	translatedLabel := c.translator(label)
	labelWidth := c.pdf.GetStringWidth(translatedLabel)

	c.pdf.SetXY(20, y)
	c.pdf.CellFormat(labelWidth, 11, translatedLabel, "", 0, "L", false, 0, urlStr)

	// Draw underline
	c.pdf.SetDrawColor(c.theme.LinkColor.R, c.theme.LinkColor.G, c.theme.LinkColor.B)
	c.pdf.SetLineWidth(0.2)
	c.pdf.Line(20, y+10, 20+labelWidth, y+10)

	// Restore normal text color
	c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)

	return y + 15
}

// renderText renders text element
func (c *Converter) renderText(text present.Text, y float64) float64 {
	// Check if this text contains markdown code blocks (```)
	content := strings.Join(text.Lines, "\n")

	// Check for markdown code block markers
	if strings.Contains(content, "```") {
		return c.renderMarkdownCodeBlock(content, y)
	}

	// Regular text rendering
	c.setTextFont("", 21)
	c.pdf.SetXY(20, y)

	// For regular text, join with spaces
	content = strings.Join(text.Lines, " ")
	c.pdf.MultiCell(257, 11, c.translator(content), "", "L", false)

	return y + 15
}

// renderMarkdownCodeBlock renders markdown code blocks (```)
func (c *Converter) renderMarkdownCodeBlock(content string, y float64) float64 {
	// Extract code block: ```language\ncode\n```
	re := regexp.MustCompile("(?s)```(\\w*)\\s*\n(.*?)```")
	match := re.FindStringSubmatch(content)

	if len(match) < 3 {
		// No valid code block found, render as plain text
		c.setTextFont("", 21)
		c.pdf.SetXY(20, y)
		c.pdf.MultiCell(257, 11, c.translator(content), "", "L", false)
		return y + 15
	}

	language := match[1]
	if language == "" {
		language = "go" // default
	}
	codeText := strings.TrimSpace(match[2])

	// Highlight the code
	tokens, err := c.highlightCode(codeText, language)
	if err != nil {
		// Fallback to plain rendering
		return c.renderCodePlain(codeText, y)
	}

	// Split tokens into lines
	lines := splitTokensIntoLines(tokens)

	// Calculate code block height
	codeHeight := float64(len(lines)) * 6
	if codeHeight > 120 {
		codeHeight = 120
	}

	// Background for code
	c.pdf.SetFillColor(c.theme.CodeBackground.R, c.theme.CodeBackground.G, c.theme.CodeBackground.B)
	c.pdf.Rect(20, y, 257, codeHeight+5, "F")

	// Render lines with syntax highlighting
	lineY := y + 2
	maxLines := 20
	for i, line := range lines {
		if i >= maxLines {
			if !c.quiet {
				fmt.Fprintf(os.Stderr, "Warning: code block truncated on slide %d \"%s\" (max %d lines, has %d)\n", c.currentSlideNumber, c.currentSlideTitle, maxLines, len(lines))
			}
			c.pdf.SetTextColor(c.theme.CodeLineNumber.R, c.theme.CodeLineNumber.G, c.theme.CodeLineNumber.B)
			c.setCodeFont("", 11)
			c.pdf.SetXY(25, lineY)
			c.pdf.Cell(0, 6, c.translator("..."))
			break
		}
		c.renderHighlightedLine(line, 25, lineY)
		lineY += 6
	}

	c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
	return y + codeHeight + 12
}

// renderList renders list element
func (c *Converter) renderList(list present.List, y float64) float64 {
	c.setTextFont("", 18)

	bullet := "• "
	for _, item := range list.Bullet {
		c.pdf.SetXY(25, y)

		fullText := bullet + item

		c.pdf.MultiCell(247, 9, c.translator(fullText), "", "L", false)
		y += 12
	}

	return y + 6
}

// renderCode renders code block
func (c *Converter) renderCode(code present.Code, y float64) float64 {
	// Extract code lines from Raw content
	codeText := string(code.Raw)

	// Detect language from filename if available
	language := "go" // default to Go
	if code.FileName != "" {
		language = detectLanguage(code.FileName)
	}

	// Highlight the code
	tokens, err := c.highlightCode(codeText, language)
	if err != nil {
		// Fallback to plain rendering if highlighting fails
		return c.renderCodePlain(codeText, y)
	}

	// Split tokens into lines
	lines := splitTokensIntoLines(tokens)

	// Calculate code block height
	codeHeight := float64(len(lines)) * 6
	if codeHeight > 120 {
		codeHeight = 120
	}

	// Background for code
	c.pdf.SetFillColor(c.theme.CodeBackground.R, c.theme.CodeBackground.G, c.theme.CodeBackground.B)
	c.pdf.Rect(20, y, 257, codeHeight+5, "F")

	// Render lines with syntax highlighting
	lineY := y + 2
	maxLines := 20
	for i, line := range lines {
		if i >= maxLines {
			if !c.quiet {
				fmt.Fprintf(os.Stderr, "Warning: code block truncated on slide %d \"%s\" (max %d lines, has %d)\n", c.currentSlideNumber, c.currentSlideTitle, maxLines, len(lines))
			}
			c.pdf.SetTextColor(c.theme.CodeLineNumber.R, c.theme.CodeLineNumber.G, c.theme.CodeLineNumber.B)
			c.setCodeFont("", 11)
			c.pdf.SetXY(25, lineY)
			c.pdf.Cell(0, 6, c.translator("..."))
			break
		}
		c.renderHighlightedLine(line, 25, lineY)
		lineY += 6
	}

	c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
	return y + codeHeight + 12
}

// extractAuthorText extracts text from author element
func (c *Converter) extractAuthorText(author present.Author) string {
	var buf bytes.Buffer
	for _, elem := range author.Elem {
		if text, ok := elem.(present.Text); ok {
			buf.WriteString(strings.Join(text.Lines, " "))
			buf.WriteString(" ")
		}
	}
	return strings.TrimSpace(buf.String())
}

// renderHTML renders HTML element (used in Markdown-enabled presentations)
func (c *Converter) renderHTML(html present.HTML, y float64) float64 {
	htmlContent := string(html.HTML)

	// Check if content contains multiple element types
	// Note: use "<pre><code" (without >) to match both <pre><code> and <pre><code class="...">
	hasCode := strings.Contains(htmlContent, "<pre><code")
	hasLists := strings.Contains(htmlContent, "<ul>") || strings.Contains(htmlContent, "<ol>")
	hasParagraphs := strings.Contains(htmlContent, "<p>")

	// Count how many different types we have
	typeCount := 0
	if hasCode {
		typeCount++
	}
	if hasLists {
		typeCount++
	}
	if hasParagraphs {
		typeCount++
	}

	// If content has multiple element types, render them in order
	if typeCount > 1 {
		return c.renderHTMLMixed(htmlContent, y)
	}

	// Handle single element types
	if hasCode {
		return c.renderHTMLCode(htmlContent, y)
	}

	if hasLists {
		return c.renderHTMLList(htmlContent, y)
	}

	if hasParagraphs {
		return c.renderHTMLParagraphs(htmlContent, y)
	}

	// Fallback: render as plain text
	return c.renderHTMLPlainText(htmlContent, y)
}

// renderHTMLParagraphs renders multiple HTML paragraphs
func (c *Converter) renderHTMLParagraphs(html string, y float64) float64 {
	// Extract all paragraphs
	re := regexp.MustCompile(`(?s)<p>(.*?)</p>`)
	matches := re.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			paragraphHTML := strings.TrimSpace(match[1])

			if paragraphHTML == "" {
				continue
			}

			// Parse HTML formatting
			fragments := parseHTMLFormatting(paragraphHTML)

			// Render formatted text
			c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
			y = c.renderFormattedText(fragments, 20, y, 257, 11)
			y += 5 // Extra spacing between paragraphs
		}
	}

	return y
}

// renderHTMLMixed renders HTML content with mixed paragraphs, lists, and code blocks in order
func (c *Converter) renderHTMLMixed(html string, y float64) float64 {
	// Split by major HTML tags while preserving them
	// Match: <p>...</p>, <ul>...</ul>, <ol>...</ol>, <pre><code>...</code></pre>
	re := regexp.MustCompile(`(?s)(<pre><code.*?</code></pre>|<p>.*?</p>|<ul>.*?</ul>|<ol>.*?</ol>)`)
	matches := re.FindAllString(html, -1)

	for _, match := range matches {
		match = strings.TrimSpace(match)
		if match == "" {
			continue
		}

		// Determine element type and render accordingly
		if strings.HasPrefix(match, "<pre><code") {
			y = c.renderHTMLCode(match, y)
		} else if strings.HasPrefix(match, "<p>") {
			y = c.renderHTMLParagraphs(match, y)
		} else if strings.HasPrefix(match, "<ul>") || strings.HasPrefix(match, "<ol>") {
			y = c.renderHTMLList(match, y)
		}
	}

	return y
}

// renderHTMLList renders HTML list
func (c *Converter) renderHTMLList(html string, y float64) float64 {
	// Extract list items
	re := regexp.MustCompile(`(?s)<li>(.*?)</li>`)
	matches := re.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			itemHTML := strings.TrimSpace(match[1])

			// Parse HTML formatting
			fragments := parseHTMLFormatting(itemHTML)

			// Render bullet
			c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
			c.setTextFont("", 18)
			c.pdf.SetXY(25, y)
			c.pdf.Cell(8, 9, c.translator("• "))

			// Render formatted text
			y = c.renderFormattedText(fragments, 30, y, 247, 9)
			y += 3
		}
	}

	return y + 6
}

// renderHTMLCode renders HTML code block
func (c *Converter) renderHTMLCode(html string, y float64) float64 {
	// Extract code content - use (?s) flag to make . match newlines
	// Updated regex to handle optional attributes in <code> tag
	re := regexp.MustCompile(`(?s)<pre><code[^>]*>(.*?)</code></pre>`)
	match := re.FindStringSubmatch(html)

	if len(match) < 2 {
		return y
	}

	codeText := match[1]
	codeText = strings.TrimSpace(codeText)

	// Decode HTML entities (e.g., &quot; -> ", &lt; -> <, etc.)
	codeText = decodeHTMLEntities(codeText)

	// Try to detect language from class attribute
	language := "go" // default
	classRe := regexp.MustCompile(`<code class="language-(\w+)">`)
	if classMatch := classRe.FindStringSubmatch(html); len(classMatch) > 1 {
		language = classMatch[1]
	}

	// Highlight the code
	tokens, err := c.highlightCode(codeText, language)
	if err != nil {
		// Fallback to plain rendering
		return c.renderCodePlain(codeText, y)
	}

	// Split tokens into lines
	lines := splitTokensIntoLines(tokens)

	// Calculate code block height
	codeHeight := float64(len(lines)) * 6
	if codeHeight > 120 {
		codeHeight = 120
	}

	// Background for code
	c.pdf.SetFillColor(c.theme.CodeBackground.R, c.theme.CodeBackground.G, c.theme.CodeBackground.B)
	c.pdf.Rect(20, y, 257, codeHeight+5, "F")

	// Render lines with syntax highlighting
	lineY := y + 2
	maxLines := 20
	for i, line := range lines {
		if i >= maxLines {
			if !c.quiet {
				fmt.Fprintf(os.Stderr, "Warning: code block truncated on slide %d \"%s\" (max %d lines, has %d)\n", c.currentSlideNumber, c.currentSlideTitle, maxLines, len(lines))
			}
			c.pdf.SetTextColor(c.theme.CodeLineNumber.R, c.theme.CodeLineNumber.G, c.theme.CodeLineNumber.B)
			c.setCodeFont("", 11)
			c.pdf.SetXY(25, lineY)
			c.pdf.Cell(0, 6, c.translator("..."))
			break
		}
		c.renderHighlightedLine(line, 25, lineY)
		lineY += 6
	}

	c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
	return y + codeHeight + 12
}

// renderHTMLPlainText renders HTML as plain text (fallback)
func (c *Converter) renderHTMLPlainText(html string, y float64) float64 {
	text := stripHTMLTags(html)
	text = strings.TrimSpace(text)

	if text == "" {
		return y
	}

	c.setTextFont("", 18)
	c.pdf.SetXY(20, y)
	c.pdf.MultiCell(257, 9, c.translator(text), "", "L", false)

	return y + 12
}

// stripHTMLTags removes HTML tags from string
func stripHTMLTags(html string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]+>`)
	text := re.ReplaceAllString(html, "")

	// Decode HTML entities
	text = decodeHTMLEntities(text)

	return text
}

// TextFragment represents a piece of text with formatting
type TextFragment struct {
	Text   string
	Bold   bool
	Italic bool
	URL    string // non-empty for clickable links
}

// parseHTMLFormatting parses HTML text and extracts fragments with formatting
func parseHTMLFormatting(html string) []TextFragment {
	var fragments []TextFragment

	// Decode HTML entities first (but not inside tags — we do it per-text-node below)
	// We process tags first, then decode entities in text nodes.

	// Regular expression to match text nodes and tags (including tags with attributes)
	re := regexp.MustCompile(`([^<]+)|(<[^>]+>)`)
	matches := re.FindAllString(html, -1)

	bold := false
	italic := false
	currentURL := ""
	var currentText strings.Builder

	flushText := func() {
		if currentText.Len() > 0 {
			text := decodeHTMLEntities(currentText.String())
			fragments = append(fragments, TextFragment{
				Text:   text,
				Bold:   bold,
				Italic: italic,
				URL:    currentURL,
			})
			currentText.Reset()
		}
	}

	// Regex to extract href from <a ...> tag
	hrefRe := regexp.MustCompile(`(?i)<a\s[^>]*href=["']([^"']+)["'][^>]*>`)

	for _, match := range matches {
		if strings.HasPrefix(match, "<") {
			flushText()

			// Process tag
			lowerMatch := strings.ToLower(match)
			switch {
			case lowerMatch == "<strong>" || lowerMatch == "<b>":
				bold = true
			case lowerMatch == "</strong>" || lowerMatch == "</b>":
				bold = false
			case lowerMatch == "<em>" || lowerMatch == "<i>":
				italic = true
			case lowerMatch == "</em>" || lowerMatch == "</i>":
				italic = false
			case strings.HasPrefix(lowerMatch, "<a "):
				if m := hrefRe.FindStringSubmatch(match); len(m) > 1 {
					currentURL = m[1]
				}
			case lowerMatch == "</a>":
				currentURL = ""
			}
		} else {
			currentText.WriteString(match)
		}
	}

	flushText()

	return fragments
}

// renderFormattedText renders text with bold, italic formatting and clickable links
// Bold/italic — visual simulation (Helvetica has no B/I variants for Cyrillic)
func (c *Converter) renderFormattedText(fragments []TextFragment, x, y, maxWidth, lineHeight float64) float64 {
	const (
		boldOffset = 0.2  // offset for bold simulation (mm)
		italicSkew = 12.0 // skew angle for italic simulation (degrees)
	)
	currentX := x
	currentY := y

	c.setTextFont("", 18)

	for _, fragment := range fragments {
		isLink := fragment.URL != ""

		if isLink {
			c.pdf.SetTextColor(c.theme.LinkColor.R, c.theme.LinkColor.G, c.theme.LinkColor.B)
		}

		words := strings.Fields(fragment.Text)
		for _, word := range words {
			translatedWord := c.translator(word + " ")
			wordWidth := c.pdf.GetStringWidth(translatedWord)

			if currentX+wordWidth > x+maxWidth && currentX > x {
				currentY += lineHeight
				currentX = x
			}

			drawWord := func() {
				c.pdf.SetXY(currentX, currentY)
				if isLink {
					// CellFormat with linkStr makes the cell area a clickable hyperlink
					c.pdf.CellFormat(wordWidth, lineHeight, translatedWord, "", 0, "L", false, 0, fragment.URL)
					// Draw underline manually
					c.pdf.SetDrawColor(c.theme.LinkColor.R, c.theme.LinkColor.G, c.theme.LinkColor.B)
					c.pdf.SetLineWidth(0.2)
					underlineY := currentY + lineHeight - 1
					c.pdf.Line(currentX, underlineY, currentX+wordWidth, underlineY)
				} else {
					c.pdf.Cell(wordWidth, lineHeight, translatedWord)
				}
			}

			if fragment.Italic {
				c.pdf.TransformBegin()
				c.pdf.TransformSkew(italicSkew, 0, currentX, currentY)
			}

			if fragment.Bold {
				drawWord()
				c.pdf.SetXY(currentX+boldOffset, currentY)
				if isLink {
					c.pdf.CellFormat(wordWidth, lineHeight, translatedWord, "", 0, "L", false, 0, fragment.URL)
				} else {
					c.pdf.Cell(wordWidth, lineHeight, translatedWord)
				}
			} else {
				drawWord()
			}

			if fragment.Italic {
				c.pdf.TransformEnd()
			}

			currentX += wordWidth
		}

		if isLink {
			// Restore normal text color
			c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
		}
	}

	return currentY + lineHeight
}

// decodeHTMLEntities decodes common HTML entities
func decodeHTMLEntities(text string) string {
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&#34;", "\"")
	text = strings.ReplaceAll(text, "&apos;", "'")
	return text
}

// highlightCode performs syntax highlighting on code
func (c *Converter) highlightCode(code, language string) ([]Token, error) {
	// Get lexer for the language
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Get style
	style := styles.Get(c.codeTheme)
	if style == nil {
		style = styles.Fallback
	}

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, err
	}

	// Convert to our Token format with colors
	var tokens []Token
	for _, token := range iterator.Tokens() {
		color := getTokenColor(token.Type, style)
		tokens = append(tokens, Token{
			Type:  token.Type,
			Value: token.Value,
			Color: color,
		})
	}

	return tokens, nil
}

// getTokenColor returns RGB color for a token type based on style
func getTokenColor(tokenType chroma.TokenType, style *chroma.Style) [3]int {
	entry := style.Get(tokenType)

	// Default color (light gray for dark background)
	defaultColor := [3]int{171, 178, 191}

	if entry.Colour.IsSet() {
		r, g, b := entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue()
		return [3]int{int(r), int(g), int(b)}
	}

	// Return color based on token type for common cases
	switch tokenType {
	case chroma.Keyword, chroma.KeywordNamespace, chroma.KeywordType:
		return [3]int{198, 120, 221} // Purple
	case chroma.String, chroma.StringDouble, chroma.StringSingle:
		return [3]int{152, 195, 121} // Green
	case chroma.Comment, chroma.CommentSingle, chroma.CommentMultiline:
		return [3]int{92, 99, 112} // Gray
	case chroma.Name, chroma.NameFunction:
		return [3]int{97, 175, 239} // Blue
	case chroma.LiteralNumber, chroma.LiteralNumberInteger, chroma.LiteralNumberFloat:
		return [3]int{209, 154, 102} // Orange
	case chroma.Operator:
		return [3]int{198, 120, 221} // Purple
	case chroma.NameBuiltin, chroma.NameClass:
		return [3]int{229, 192, 123} // Yellow
	default:
		return defaultColor
	}
}

// splitTokensIntoLines splits tokens into lines
func splitTokensIntoLines(tokens []Token) [][]Token {
	if len(tokens) == 0 {
		return [][]Token{}
	}

	var lines [][]Token
	var currentLine []Token

	for _, token := range tokens {
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				// New line encountered - append current line (even if empty)
				lines = append(lines, currentLine)
				currentLine = nil
			}
			if part != "" {
				currentLine = append(currentLine, Token{
					Type:  token.Type,
					Value: part,
					Color: token.Color,
				})
			}
		}
	}

	// Add the last line
	lines = append(lines, currentLine)

	return lines
}

// renderHighlightedLine renders a line of syntax-highlighted tokens
func (c *Converter) renderHighlightedLine(tokens []Token, x, y float64) {
	currentX := x

	for _, token := range tokens {
		c.pdf.SetTextColor(token.Color[0], token.Color[1], token.Color[2])
		c.pdf.SetXY(currentX, y)

		// Translate token value for UTF-8 support
		value := c.translator(token.Value)

		// Use JetBrains Mono for code - monospace font with Cyrillic support
		c.setCodeFont("", 11)

		// Get width of the text to advance X position
		width := c.pdf.GetStringWidth(value)
		c.pdf.Cell(width, 6, value)

		currentX += width
	}
}

// detectLanguage detects programming language from filename
func detectLanguage(filename string) string {
	ext := ""
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		ext = filename[idx+1:]
	}

	switch ext {
	case "go":
		return "go"
	case "py":
		return "python"
	case "js":
		return "javascript"
	case "ts":
		return "typescript"
	case "java":
		return "java"
	case "c":
		return "c"
	case "cpp", "cc", "cxx":
		return "cpp"
	case "rs":
		return "rust"
	case "rb":
		return "ruby"
	case "php":
		return "php"
	case "sh", "bash":
		return "bash"
	case "html":
		return "html"
	case "css":
		return "css"
	case "json":
		return "json"
	case "xml":
		return "xml"
	case "yaml", "yml":
		return "yaml"
	case "sql":
		return "sql"
	default:
		return "go" // default to Go
	}
}

// renderCodePlain renders code without syntax highlighting (fallback)
func (c *Converter) renderCodePlain(code string, y float64) float64 {
	lines := strings.Split(code, "\n")

	// Background for code
	c.pdf.SetFillColor(c.theme.CodeBackground.R, c.theme.CodeBackground.G, c.theme.CodeBackground.B)
	codeHeight := float64(len(lines)) * 6
	if codeHeight > 120 {
		codeHeight = 120
	}

	c.pdf.Rect(20, y, 257, codeHeight+5, "F")

	// Code text - use JetBrains Mono for monospace with Cyrillic support
	c.setCodeFont("", 11)
	c.pdf.SetTextColor(c.theme.CodeText.R, c.theme.CodeText.G, c.theme.CodeText.B)

	lineY := y + 2
	maxLines := 20
	for i, line := range lines {
		if i >= maxLines {
			if !c.quiet {
				fmt.Fprintf(os.Stderr, "Warning: code block truncated on slide %d \"%s\" (max %d lines, has %d)\n", c.currentSlideNumber, c.currentSlideTitle, maxLines, len(lines))
			}
			c.pdf.SetXY(25, lineY)
			c.pdf.Cell(0, 6, c.translator("..."))
			break
		}
		c.pdf.SetXY(25, lineY)
		c.pdf.Cell(0, 6, c.translator(line))
		lineY += 6
	}

	c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
	return y + codeHeight + 12
}
