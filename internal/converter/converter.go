package converter

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/tools/present"
)

// Converter handles conversion from .slide to PDF
type Converter struct {
	pdf *gofpdf.Fpdf
}

// NewConverter creates a new converter instance
func NewConverter() *Converter {
	return &Converter{}
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

	// Create PDF
	c.pdf = gofpdf.New("L", "mm", "A4", "")
	c.pdf.SetAutoPageBreak(false, 0)

	// Render title slide
	c.renderTitleSlide(doc)

	// Render each section as a slide
	for _, section := range doc.Sections {
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
	c.pdf.SetFillColor(41, 128, 185)
	c.pdf.Rect(0, 0, 297, 210, "F")

	// Title
	c.pdf.SetTextColor(255, 255, 255)
	c.pdf.SetFont("Arial", "B", 36)
	c.pdf.SetXY(20, 70)
	c.pdf.MultiCell(257, 15, doc.Title, "", "C", false)

	// Subtitle
	if doc.Subtitle != "" {
		c.pdf.SetFont("Arial", "", 20)
		c.pdf.SetXY(20, 95)
		c.pdf.MultiCell(257, 10, doc.Subtitle, "", "C", false)
	}

	// Authors
	if len(doc.Authors) > 0 {
		c.pdf.SetFont("Arial", "", 14)
		y := 130.0
		for _, author := range doc.Authors {
			authorText := c.extractAuthorText(author)
			if authorText != "" {
				c.pdf.SetXY(20, y)
				c.pdf.MultiCell(257, 8, authorText, "", "C", false)
				y += 10
			}
		}
	}

	// Date
	if !doc.Time.IsZero() {
		c.pdf.SetFont("Arial", "I", 12)
		c.pdf.SetXY(20, 180)
		c.pdf.MultiCell(257, 6, doc.Time.Format("January 2, 2006"), "", "C", false)
	}
}

// renderSlide renders a single slide
func (c *Converter) renderSlide(section present.Section) {
	c.pdf.AddPage()

	// Background
	c.pdf.SetFillColor(255, 255, 255)
	c.pdf.Rect(0, 0, 297, 210, "F")

	// Title
	c.pdf.SetTextColor(41, 128, 185)
	c.pdf.SetFont("Arial", "B", 24)
	c.pdf.SetXY(20, 15)
	c.pdf.MultiCell(257, 10, section.Title, "", "L", false)

	// Draw a line under the title
	c.pdf.SetDrawColor(41, 128, 185)
	c.pdf.SetLineWidth(0.5)
	c.pdf.Line(20, 30, 277, 30)

	// Content
	c.pdf.SetTextColor(0, 0, 0)
	y := 40.0

	for _, elem := range section.Elem {
		y = c.renderElement(elem, y)
		if y > 190 {
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
	default:
		// Skip unsupported elements
		return y
	}
}

// renderText renders text element
func (c *Converter) renderText(text present.Text, y float64) float64 {
	c.pdf.SetFont("Arial", "", 14)
	c.pdf.SetXY(20, y)

	content := strings.Join(text.Lines, " ")
	c.pdf.MultiCell(257, 7, content, "", "L", false)

	return y + 10
}

// renderList renders list element
func (c *Converter) renderList(list present.List, y float64) float64 {
	c.pdf.SetFont("Arial", "", 12)

	for _, item := range list.Bullet {
		c.pdf.SetXY(25, y)

		// Bullet point
		bullet := "-"

		fullText := bullet + " " + item

		c.pdf.MultiCell(247, 6, fullText, "", "L", false)
		y += 8
	}

	return y + 4
}

// renderCode renders code block
func (c *Converter) renderCode(code present.Code, y float64) float64 {
	// Extract code lines from Raw content
	lines := strings.Split(string(code.Raw), "\n")

	// Background for code
	c.pdf.SetFillColor(240, 240, 240)
	codeHeight := float64(len(lines)) * 5
	if codeHeight > 80 {
		codeHeight = 80
	}

	c.pdf.Rect(20, y, 257, codeHeight+4, "F")

	// Code text
	c.pdf.SetFont("Courier", "", 10)
	c.pdf.SetTextColor(0, 0, 0)

	lineY := y + 2
	maxLines := 12
	for i, line := range lines {
		if i >= maxLines {
			c.pdf.SetXY(25, lineY)
			c.pdf.Cell(0, 5, "...")
			break
		}
		c.pdf.SetXY(25, lineY)
		c.pdf.Cell(0, 5, line)
		lineY += 5
	}

	c.pdf.SetTextColor(0, 0, 0)
	return y + codeHeight + 10
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

	// Handle code blocks first (most specific)
	if strings.Contains(htmlContent, "<pre><code>") {
		return c.renderHTMLCode(htmlContent, y)
	}

	// Handle lists
	if strings.Contains(htmlContent, "<ul>") || strings.Contains(htmlContent, "<ol>") {
		return c.renderHTMLList(htmlContent, y)
	}

	// Handle paragraphs (may contain multiple <p> tags)
	if strings.Contains(htmlContent, "<p>") {
		return c.renderHTMLParagraphs(htmlContent, y)
	}

	// Fallback: render as plain text
	return c.renderHTMLPlainText(htmlContent, y)
}

// renderHTMLParagraphs renders multiple HTML paragraphs
func (c *Converter) renderHTMLParagraphs(html string, y float64) float64 {
	// Extract all paragraphs
	re := regexp.MustCompile(`<p>(.*?)</p>`)
	matches := re.FindAllStringSubmatch(html, -1)

	c.pdf.SetFont("Arial", "", 14)

	for _, match := range matches {
		if len(match) > 1 {
			text := stripHTMLTags(match[1])
			text = strings.TrimSpace(text)

			if text == "" {
				continue
			}

			c.pdf.SetXY(20, y)
			c.pdf.MultiCell(257, 7, text, "", "L", false)
			y += 10
		}
	}

	return y
}

// renderHTMLList renders HTML list
func (c *Converter) renderHTMLList(html string, y float64) float64 {
	c.pdf.SetFont("Arial", "", 12)

	// Extract list items
	re := regexp.MustCompile(`<li>(.*?)</li>`)
	matches := re.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			item := stripHTMLTags(match[1])
			item = strings.TrimSpace(item)

			c.pdf.SetXY(25, y)
			c.pdf.MultiCell(247, 6, "- "+item, "", "L", false)
			y += 8
		}
	}

	return y + 4
}

// renderHTMLCode renders HTML code block
func (c *Converter) renderHTMLCode(html string, y float64) float64 {
	// Extract code content
	re := regexp.MustCompile(`<pre><code>(.*?)</code></pre>`)
	match := re.FindStringSubmatch(html)

	if len(match) < 2 {
		return y
	}

	code := match[1]
	code = strings.TrimSpace(code)
	lines := strings.Split(code, "\n")

	// Background for code
	c.pdf.SetFillColor(240, 240, 240)
	codeHeight := float64(len(lines)) * 5
	if codeHeight > 80 {
		codeHeight = 80
	}

	c.pdf.Rect(20, y, 257, codeHeight+4, "F")

	// Code text
	c.pdf.SetFont("Courier", "", 10)
	c.pdf.SetTextColor(0, 0, 0)

	lineY := y + 2
	maxLines := 12
	for i, line := range lines {
		if i >= maxLines {
			c.pdf.SetXY(25, lineY)
			c.pdf.Cell(0, 5, "...")
			break
		}
		c.pdf.SetXY(25, lineY)
		c.pdf.Cell(0, 5, line)
		lineY += 5
	}

	c.pdf.SetTextColor(0, 0, 0)
	return y + codeHeight + 10
}

// renderHTMLPlainText renders HTML as plain text (fallback)
func (c *Converter) renderHTMLPlainText(html string, y float64) float64 {
	text := stripHTMLTags(html)
	text = strings.TrimSpace(text)

	if text == "" {
		return y
	}

	c.pdf.SetFont("Arial", "", 12)
	c.pdf.SetXY(20, y)
	c.pdf.MultiCell(257, 6, text, "", "L", false)

	return y + 8
}

// stripHTMLTags removes HTML tags from string
func stripHTMLTags(html string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]+>`)
	text := re.ReplaceAllString(html, "")

	// Decode HTML entities
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	return text
}
