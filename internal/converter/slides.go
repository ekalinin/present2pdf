package converter

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/present"
)

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
