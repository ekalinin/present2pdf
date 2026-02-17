package converter

import (
	"strings"

	"golang.org/x/tools/present"
)

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
