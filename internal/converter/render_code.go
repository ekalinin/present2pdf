package converter

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"golang.org/x/tools/present"
)

// Token represents a syntax-highlighted token
type Token struct {
	Type  chroma.TokenType
	Value string
	Color [3]int // RGB color
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

	return c.renderHighlightedCode(tokens, y)
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

	return c.renderHighlightedCode(tokens, y)
}

// renderHighlightedCode renders syntax-highlighted tokens as a code block
func (c *Converter) renderHighlightedCode(tokens []Token, y float64) float64 {
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
