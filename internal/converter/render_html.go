package converter

import (
	"regexp"
	"strings"

	"golang.org/x/tools/present"
)

// TextFragment represents a piece of text with formatting
type TextFragment struct {
	Text   string
	Bold   bool
	Italic bool
	Code   bool   // inline code (monospace font + background)
	URL    string // non-empty for clickable links
}

// renderHTML renders HTML element (used in Markdown-enabled presentations)
func (c *Converter) renderHTML(html present.HTML, y float64) float64 {
	htmlContent := string(html.HTML)

	// Check if content contains multiple element types
	// Note: use "<pre><code" (without >) to match both <pre><code> and <pre><code class="...">
	hasCode := strings.Contains(htmlContent, "<pre><code")
	hasLists := strings.Contains(htmlContent, "<ul>") || strings.Contains(htmlContent, "<ol>")
	hasParagraphs := strings.Contains(htmlContent, "<p>")
	hasBlockquote := strings.Contains(htmlContent, "<blockquote>")

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
	if hasBlockquote {
		typeCount++
	}

	// If content has multiple element types, render them in order
	if typeCount > 1 {
		return c.renderHTMLMixed(htmlContent, y)
	}

	// Handle single element types
	if hasBlockquote {
		return c.renderHTMLBlockquote(htmlContent, y)
	}

	if hasCode {
		return c.renderHTMLCode(htmlContent, y)
	}

	if hasLists {
		return c.renderHTMLList(htmlContent, y)
	}

	if hasParagraphs {
		return c.renderHTMLParagraphs(htmlContent, y)
	}

	// Standalone <img> tag (not wrapped in <p>)
	if strings.Contains(htmlContent, "<img ") {
		return c.renderHTMLImage(htmlContent, y)
	}

	// Fallback: render as plain text
	return c.renderHTMLPlainText(htmlContent, y)
}

// renderHTMLMixed renders HTML content with mixed paragraphs, lists, code blocks, and blockquotes in order
func (c *Converter) renderHTMLMixed(html string, y float64) float64 {
	// Split by major HTML tags while preserving them
	// Blockquote is listed first to take priority over inner <p> tags
	re := regexp.MustCompile(`(?s)(<blockquote>.*?</blockquote>|<pre><code.*?</code></pre>|<p>.*?</p>|<ul>.*?</ul>|<ol>.*?</ol>)`)
	matches := re.FindAllString(html, -1)

	for _, match := range matches {
		match = strings.TrimSpace(match)
		if match == "" {
			continue
		}

		// Determine element type and render accordingly
		if strings.HasPrefix(match, "<blockquote>") {
			y = c.renderHTMLBlockquote(match, y)
		} else if strings.HasPrefix(match, "<pre><code") {
			y = c.renderHTMLCode(match, y)
		} else if strings.HasPrefix(match, "<p>") {
			y = c.renderHTMLParagraphs(match, y)
		} else if strings.HasPrefix(match, "<ul>") || strings.HasPrefix(match, "<ol>") {
			y = c.renderHTMLList(match, y)
		}
	}

	return y
}

// renderHTMLParagraphs renders multiple HTML paragraphs
func (c *Converter) renderHTMLParagraphs(html string, y float64) float64 {
	// Extract all paragraphs
	re := regexp.MustCompile(`(?s)<p>(.*?)</p>`)
	matches := re.FindAllStringSubmatch(html, -1)

	imgTagRe := regexp.MustCompile(`(?i)^<img\s`)

	for _, match := range matches {
		if len(match) > 1 {
			paragraphHTML := strings.TrimSpace(match[1])

			if paragraphHTML == "" {
				continue
			}

			// Paragraph contains only an image tag — render as image
			if imgTagRe.MatchString(paragraphHTML) {
				y = c.renderHTMLImage(paragraphHTML, y)
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

	codeText := strings.TrimSpace(match[1])

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

	return c.renderHighlightedCode(tokens, y)
}

// renderHTMLBlockquote renders a Markdown blockquote (> text) as a styled block
func (c *Converter) renderHTMLBlockquote(html string, y float64) float64 {
	re := regexp.MustCompile(`(?s)<blockquote>\s*(.*?)\s*</blockquote>`)
	match := re.FindStringSubmatch(html)
	if len(match) < 2 {
		return y
	}
	inner := strings.TrimSpace(match[1])

	// Extract paragraphs from inner content
	paraRe := regexp.MustCompile(`(?s)<p>(.*?)</p>`)
	paraMatches := paraRe.FindAllStringSubmatch(inner, -1)

	var paragraphsHTML []string
	if len(paraMatches) == 0 {
		// No <p> tags — treat whole inner content as one paragraph
		text := stripHTMLTags(inner)
		if t := strings.TrimSpace(text); t != "" {
			paragraphsHTML = []string{t}
		}
	} else {
		for _, m := range paraMatches {
			if len(m) > 1 {
				if t := strings.TrimSpace(m[1]); t != "" {
					paragraphsHTML = append(paragraphsHTML, t)
				}
			}
		}
	}

	if len(paragraphsHTML) == 0 {
		return y
	}

	const (
		borderWidth = 4.0  // mm
		textX       = 28.0 // absolute X for text start (after left border)
		textWidth   = 249.0
		lineHeight  = 11.0
		paddingV    = 4.0 // vertical padding top and bottom
		paraSpacing = 3.0 // spacing between paragraphs
	)

	// Estimate total height using font metrics
	c.setTextFont("", 18)
	totalHeight := paddingV * 2
	for i, paraHTML := range paragraphsHTML {
		plainText := stripHTMLTags(paraHTML)
		words := strings.Fields(plainText)
		lineWidth := 0.0
		lines := 1
		for _, word := range words {
			ww := c.pdf.GetStringWidth(c.translator(word + " "))
			if lineWidth+ww > textWidth && lineWidth > 0 {
				lines++
				lineWidth = ww
			} else {
				lineWidth += ww
			}
		}
		totalHeight += float64(lines) * lineHeight
		if i < len(paragraphsHTML)-1 {
			totalHeight += paraSpacing
		}
	}

	// Draw background rectangle
	c.pdf.SetFillColor(c.theme.BlockquoteBackground.R, c.theme.BlockquoteBackground.G, c.theme.BlockquoteBackground.B)
	c.pdf.Rect(20, y, 257, totalHeight, "F")

	// Draw left border
	c.pdf.SetFillColor(c.theme.BlockquoteBorder.R, c.theme.BlockquoteBorder.G, c.theme.BlockquoteBorder.B)
	c.pdf.Rect(20, y, borderWidth, totalHeight, "F")

	// Render paragraph text on top
	textY := y + paddingV
	for i, paraHTML := range paragraphsHTML {
		fragments := parseHTMLFormatting(paraHTML)
		c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
		textY = c.renderFormattedText(fragments, textX, textY, textWidth, lineHeight)
		if i < len(paragraphsHTML)-1 {
			textY += paraSpacing
		}
	}

	return y + totalHeight + 5
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
	code := false
	currentURL := ""
	var currentText strings.Builder

	flushText := func() {
		if currentText.Len() > 0 {
			text := decodeHTMLEntities(currentText.String())
			fragments = append(fragments, TextFragment{
				Text:   text,
				Bold:   bold,
				Italic: italic,
				Code:   code,
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
			case lowerMatch == "<code>":
				code = true
			case lowerMatch == "</code>":
				code = false
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
		isCode := fragment.Code

		if isCode {
			c.setCodeFont("", 16)
			c.pdf.SetTextColor(c.theme.InlineCodeText.R, c.theme.InlineCodeText.G, c.theme.InlineCodeText.B)
		} else if isLink {
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

			if isCode {
				c.pdf.SetFillColor(c.theme.InlineCodeBackground.R, c.theme.InlineCodeBackground.G, c.theme.InlineCodeBackground.B)
				c.pdf.Rect(currentX, currentY+0.5, wordWidth, lineHeight-1, "F")
				c.pdf.SetTextColor(c.theme.InlineCodeText.R, c.theme.InlineCodeText.G, c.theme.InlineCodeText.B)
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

		if isCode {
			c.setTextFont("", 18)
			c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
		} else if isLink {
			// Restore normal text color
			c.pdf.SetTextColor(c.theme.SlideText.R, c.theme.SlideText.G, c.theme.SlideText.B)
		}
	}

	return currentY + lineHeight
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
