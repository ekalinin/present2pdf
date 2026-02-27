package converter

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/jung-kurt/gofpdf"
	"golang.org/x/tools/present"
)

func TestNewConverter(t *testing.T) {
	conv := NewConverter()
	if conv == nil {
		t.Error("NewConverter() returned nil")
	}
}

func TestExtractAuthorText(t *testing.T) {
	tests := []struct {
		name     string
		author   present.Author
		expected string
	}{
		{
			name: "single text element",
			author: present.Author{
				Elem: []present.Elem{
					present.Text{Lines: []string{"John Doe"}},
				},
			},
			expected: "John Doe",
		},
		{
			name: "multiple text elements",
			author: present.Author{
				Elem: []present.Elem{
					present.Text{Lines: []string{"John Doe"}},
					present.Text{Lines: []string{"john@example.com"}},
				},
			},
			expected: "John Doe john@example.com",
		},
		{
			name: "multiline text",
			author: present.Author{
				Elem: []present.Elem{
					present.Text{Lines: []string{"John", "Doe"}},
				},
			},
			expected: "John Doe",
		},
	}

	conv := NewConverter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.extractAuthorText(tt.author)
			if result != tt.expected {
				t.Errorf("extractAuthorText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestConvertBasic(t *testing.T) {
	// Create temporary slide file
	slideContent := `Test Presentation
Test Subtitle
15 Feb 2026

Test Author
test@example.com

* First Slide

This is test content.

- Bullet 1
- Bullet 2

* Second Slide

	code example
	line 2
`

	tmpFile, err := os.CreateTemp("", "test-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Create output file path
	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	// Convert
	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() error = %v", err)
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created")
	}
}

func TestConvertNonexistentFile(t *testing.T) {
	conv := NewConverter()
	err := conv.Convert("nonexistent.slide", "output.pdf")
	if err == nil {
		t.Error("Convert() expected error for nonexistent file, got nil")
	}
}

func TestConvertInvalidSlideFile(t *testing.T) {
	// Create temporary invalid slide file
	tmpFile, err := os.CreateTemp("", "invalid-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid content
	if _, err := tmpFile.Write([]byte("Invalid content\nwithout proper format")); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err == nil {
		t.Error("Convert() expected error for invalid slide file, got nil")
	}
}

func TestConvertMarkdownEnabled(t *testing.T) {
	// Test Markdown-enabled format (with "# " prefix)
	slideContent := `# Markdown Test Presentation
Subtitle for Markdown
15 Feb 2026

Author Name
author@example.com

## First Slide

This is _italic_ and this is **bold** text.

- First bullet
- Second bullet

## Code Slide

Example code:

	package main
	
	func main() {
		fmt.Println("Hello")
	}

## Links Slide

Visit [Go website](https://golang.org) for more info.

// This is a comment - should be ignored
`

	tmpFile, err := os.CreateTemp("", "markdown-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	// Convert
	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() failed for Markdown-enabled format: %v", err)
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created for Markdown format")
	}

	// Check file size is reasonable (should be > 1KB)
	info, _ := os.Stat(outputPath)
	if info.Size() < 1024 {
		t.Errorf("PDF file too small: %d bytes", info.Size())
	}
}

func TestConvertLegacyFormat(t *testing.T) {
	// Test legacy format (without "# " prefix)
	slideContent := `Legacy Test Presentation
Legacy Subtitle
15 Feb 2026

Author Name
author@example.com

* First Slide

This is _italic_ and this is *bold* text.

- First bullet
- Second bullet

* Code Slide

Example code:

	package main
	
	func main() {
		fmt.Println("Hello")
	}
`

	tmpFile, err := os.CreateTemp("", "legacy-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	// Convert
	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() failed for legacy format: %v", err)
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created for legacy format")
	}
}

func TestMarkdownSyntaxElements(t *testing.T) {
	// Test various Markdown syntax elements
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "subsections with ###",
			content: `# Test
15 Feb 2026

Author

## Main Section

### Subsection

Content here.
`,
			wantErr: false,
		},
		{
			name: "anchor IDs",
			content: `# Test
15 Feb 2026

Author

## Section Title {#custom-anchor}

Content with custom anchor.
`,
			wantErr: false,
		},
		{
			name: "comments with //",
			content: `# Test
15 Feb 2026

Author

## Slide

// This is a comment
Regular text here.
// Another comment
`,
			wantErr: false,
		},
		{
			name: "markdown links",
			content: `# Test
15 Feb 2026

Author

## Links

Visit [Google](https://google.com) and [GitHub](https://github.com).
`,
			wantErr: false,
		},
		{
			name: "complex formatting",
			content: `# Test
15 Feb 2026

Author

## Formatting

This is _italic_, **bold**, and ` + "`code`" + `.

- List item with _emphasis_
- List item with **bold**
- List item with ` + "`inline code`" + `
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "syntax-*.slide")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write([]byte(tt.content)); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tmpFile.Close()

			outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
			defer os.Remove(outputPath)

			conv := NewConverter()
			err = conv.Convert(tmpFile.Name(), outputPath)

			if tt.wantErr && err == nil {
				t.Errorf("Convert() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Convert() unexpected error: %v", err)
			}
			if !tt.wantErr {
				// Check if output file exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("Output PDF file was not created")
				}
			}
		})
	}
}

func TestMarkdownVsLegacyFormatDetection(t *testing.T) {
	tests := []struct {
		name       string
		firstLine  string
		isMarkdown bool
	}{
		{
			name:       "markdown format with # prefix",
			firstLine:  "# My Presentation",
			isMarkdown: true,
		},
		{
			name:       "legacy format without # prefix",
			firstLine:  "My Presentation",
			isMarkdown: false,
		},
		{
			name:       "markdown with leading space (invalid)",
			firstLine:  " # My Presentation",
			isMarkdown: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple check: does it start with "# "
			hasMarkdownPrefix := strings.HasPrefix(tt.firstLine, "# ")

			if hasMarkdownPrefix != tt.isMarkdown {
				t.Errorf("Format detection failed: got %v, want %v for %q",
					hasMarkdownPrefix, tt.isMarkdown, tt.firstLine)
			}
		})
	}
}

func TestConvertWithMultipleSections(t *testing.T) {
	// Test presentation with multiple sections and subsections
	slideContent := `# Multi-Section Presentation
Testing multiple sections
15 Feb 2026

Test Author
test@example.com

## Section 1

Content for section 1.

### Subsection 1.1

Subsection content.

### Subsection 1.2

More subsection content.

## Section 2

Content for section 2.

- Point one
- Point two
- Point three

## Section 3

	code block
	with multiple
	lines

## Conclusion

Final slide.
`

	tmpFile, err := os.CreateTemp("", "multi-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() failed for multi-section presentation: %v", err)
	}

	// Check if output file exists and has reasonable size
	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created")
	}
	if info.Size() < 2048 {
		t.Errorf("PDF file too small for multi-section presentation: %d bytes", info.Size())
	}
}

func TestHighlightCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		language string
		wantErr  bool
	}{
		{
			name: "simple go code",
			code: `package main

func main() {
	fmt.Println("Hello")
}`,
			language: "go",
			wantErr:  false,
		},
		{
			name:     "empty code",
			code:     "",
			language: "go",
			wantErr:  false,
		},
		{
			name:     "single line",
			code:     `fmt.Println("test")`,
			language: "go",
			wantErr:  false,
		},
		{
			name: "python code",
			code: `def hello():
    print("Hello")`,
			language: "python",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConverter()
			tokens, err := conv.highlightCode(tt.code, tt.language)
			if (err != nil) != tt.wantErr {
				t.Errorf("highlightCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(tokens) == 0 && tt.code != "" {
				t.Errorf("highlightCode() returned empty tokens for non-empty code")
			}
		})
	}
}

func TestSplitTokensIntoLines(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []Token
		wantLines int
	}{
		{
			name: "single line",
			tokens: []Token{
				{Value: "package", Color: [3]int{198, 120, 221}},
				{Value: " ", Color: [3]int{171, 178, 191}},
				{Value: "main", Color: [3]int{171, 178, 191}},
			},
			wantLines: 1,
		},
		{
			name: "multiple lines",
			tokens: []Token{
				{Value: "package", Color: [3]int{198, 120, 221}},
				{Value: " ", Color: [3]int{171, 178, 191}},
				{Value: "main", Color: [3]int{171, 178, 191}},
				{Value: "\n", Color: [3]int{171, 178, 191}},
				{Value: "func", Color: [3]int{198, 120, 221}},
				{Value: " ", Color: [3]int{171, 178, 191}},
				{Value: "main", Color: [3]int{171, 178, 191}},
			},
			wantLines: 2,
		},
		{
			name: "empty lines",
			tokens: []Token{
				{Value: "line1", Color: [3]int{171, 178, 191}},
				{Value: "\n\n", Color: [3]int{171, 178, 191}},
				{Value: "line3", Color: [3]int{171, 178, 191}},
			},
			wantLines: 3,
		},
		{
			name:      "empty tokens",
			tokens:    []Token{},
			wantLines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := splitTokensIntoLines(tt.tokens)
			if len(lines) != tt.wantLines {
				t.Errorf("splitTokensIntoLines() got %d lines, want %d lines", len(lines), tt.wantLines)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"go file", "main.go", "go"},
		{"python file", "script.py", "python"},
		{"javascript file", "app.js", "javascript"},
		{"typescript file", "app.ts", "typescript"},
		{"c file", "program.c", "c"},
		{"cpp file", "program.cpp", "cpp"},
		{"rust file", "main.rs", "rust"},
		{"no extension", "README", "go"},
		{"empty filename", "", "go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectLanguage(tt.filename)
			if got != tt.want {
				t.Errorf("detectLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestRenderCodeWithSyntaxHighlighting(t *testing.T) {
	// Test that code rendering with syntax highlighting works
	slideContent := `# Syntax Highlighting Test
Test Code Rendering
15 Feb 2026

Test Author

## Go Code

Simple Go example:

	package main
	
	import "fmt"
	
	func main() {
		fmt.Println("Hello, World!")
	}

## Python Code

Python example:

	def greet(name):
		return f"Hello, {name}!"
	
	print(greet("World"))
`

	tmpFile, err := os.CreateTemp("", "syntax-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() failed for syntax highlighting: %v", err)
	}

	// Check if output file exists and has reasonable size
	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created")
	}
	if info.Size() < 2048 {
		t.Errorf("PDF file too small: %d bytes (expected > 2048)", info.Size())
	}
}

func TestRenderCodePlain(t *testing.T) {
	// Test fallback to plain rendering
	conv := NewConverter()
	conv.pdf = gofpdf.New("L", "mm", "A4", "")
	conv.pdf.AddPage()
	// Initialize translator for UTF-8 support (cp1251 for Cyrillic)
	conv.translator = conv.pdf.UnicodeTranslatorFromDescriptor("cp1251")

	y := conv.renderCodePlain("test code\nline 2", 40.0)

	if y <= 40.0 {
		t.Errorf("renderCodePlain() did not advance Y position")
	}
}

func TestHighlightCodeDebug(t *testing.T) {
	// Debug test to see what tokens are generated
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}`

	conv := NewConverter()
	tokens, err := conv.highlightCode(code, "go")
	if err != nil {
		t.Fatalf("highlightCode() error = %v", err)
	}

	t.Logf("Total tokens: %d", len(tokens))
	for i, token := range tokens {
		t.Logf("Token %d: Type=%v, Value=%q, Color=%v", i, token.Type, token.Value, token.Color)
	}

	lines := splitTokensIntoLines(tokens)
	t.Logf("Total lines: %d", len(lines))
	for i, line := range lines {
		t.Logf("Line %d: %d tokens", i, len(line))
		for j, token := range line {
			t.Logf("  Token %d: %q", j, token.Value)
		}
	}
}

func TestGetTokenColor(t *testing.T) {
	// Test that getTokenColor returns valid RGB values
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	tests := []chroma.TokenType{
		chroma.Keyword,
		chroma.String,
		chroma.Comment,
		chroma.Name,
		chroma.LiteralNumber,
		chroma.Operator,
		chroma.NameBuiltin,
	}

	for _, tokenType := range tests {
		color := getTokenColor(tokenType, style)
		// Check that RGB values are in valid range
		if color[0] < 0 || color[0] > 255 ||
			color[1] < 0 || color[1] > 255 ||
			color[2] < 0 || color[2] > 255 {
			t.Errorf("getTokenColor() returned invalid RGB: %v for type %v", color, tokenType)
		}
	}
}

func TestRenderCodeWithEmptyLines(t *testing.T) {
	// Test code with empty lines
	slideContent := `# Empty Lines Test
Test
16 Feb 2026

Author

## Code with Empty Lines

	line1
	
	line3
	
	
	line6
`

	tmpFile, err := os.CreateTemp("", "empty-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() failed for code with empty lines: %v", err)
	}

	// Check if output file exists
	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created")
	}
	if info.Size() < 1024 {
		t.Errorf("PDF file too small: %d bytes", info.Size())
	}
}

func TestRenderCodeWithSpecialCharacters(t *testing.T) {
	// Test code with special characters
	slideContent := `# Special Characters Test
Test
16 Feb 2026

Author

## Code with Special Chars

	str := "Hello \"World\""
	regex := /\d+/
	path := C:\Users\test
	arrow := ->
`

	tmpFile, err := os.CreateTemp("", "special-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() failed for code with special characters: %v", err)
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created")
	}
}

func TestRenderCodeWithLongLines(t *testing.T) {
	// Test code with very long lines
	slideContent := `# Long Lines Test
Test
16 Feb 2026

Author

## Code with Long Line

	verylongvariablename := "This is a very long string that should be handled properly by the PDF renderer without causing issues"
`

	tmpFile, err := os.CreateTemp("", "longline-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() failed for code with long lines: %v", err)
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created")
	}
}

func TestHTMLCodeRegexMultiline(t *testing.T) {
	// Test that regex properly captures multiline code
	html := `<pre><code>line1
line2
line3</code></pre>`

	re := regexp.MustCompile(`(?s)<pre><code>(.*?)</code></pre>`)
	match := re.FindStringSubmatch(html)

	if len(match) < 2 {
		t.Fatal("Regex didn't match multiline HTML code")
	}

	expected := "line1\nline2\nline3"
	actual := strings.TrimSpace(match[1])

	if actual != expected {
		t.Errorf("Regex extracted %q, want %q", actual, expected)
	}
}

func TestDecodeHTMLEntities(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double quotes",
			input:    `fmt.Println(&quot;Hello&quot;)`,
			expected: `fmt.Println("Hello")`,
		},
		{
			name:     "less than and greater than",
			input:    `if x &lt; 10 &amp;&amp; y &gt; 5`,
			expected: `if x < 10 && y > 5`,
		},
		{
			name:     "ampersand",
			input:    `a &amp; b`,
			expected: `a & b`,
		},
		{
			name:     "single quote",
			input:    `char c = &#39;x&#39;`,
			expected: `char c = 'x'`,
		},
		{
			name:     "numeric quote",
			input:    `str := &#34;test&#34;`,
			expected: `str := "test"`,
		},
		{
			name:     "apostrophe",
			input:    `don&apos;t`,
			expected: `don't`,
		},
		{
			name:     "mixed entities",
			input:    `fmt.Printf(&quot;x &lt; %d&quot;, &amp;val)`,
			expected: `fmt.Printf("x < %d", &val)`,
		},
		{
			name:     "no entities",
			input:    `fmt.Println("test")`,
			expected: `fmt.Println("test")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeHTMLEntities(tt.input)
			if result != tt.expected {
				t.Errorf("decodeHTMLEntities(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRenderCodeWithHTMLEntities(t *testing.T) {
	// Test that HTML entities are properly decoded in code blocks
	slideContent := `# HTML Entities Test
Test
16 Feb 2026

Author

## Code with Quotes

	str := "Hello World"
	fmt.Printf("Value: %d", 42)
`

	tmpFile, err := os.CreateTemp("", "entities-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() failed for code with HTML entities: %v", err)
	}

	// Check if output file exists and has reasonable size
	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created")
	}
	if info.Size() < 1024 {
		t.Errorf("PDF file too small: %d bytes", info.Size())
	}
}

func TestGetAvailableStyles(t *testing.T) {
	styles := GetAvailableStyles()

	if len(styles) == 0 {
		t.Error("GetAvailableStyles() returned empty list")
	}

	// Check for some common styles
	hasMonokai := false
	for _, style := range styles {
		if style == "monokai" {
			hasMonokai = true
			break
		}
	}

	if !hasMonokai {
		t.Error("GetAvailableStyles() should include 'monokai' style")
	}

	t.Logf("Available styles: %v", styles)
}

func TestNewConverterWithCodeTheme(t *testing.T) {
	tests := []struct {
		name      string
		codeTheme string
	}{
		{"monokai style", "monokai"},
		{"github style", "github"},
		{"dracula style", "dracula"},
		{"vim style", "vim"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConverter(WithCodeTheme(tt.codeTheme))
			if conv == nil {
				t.Error("NewConverter() returned nil")
			}
			if conv.codeTheme != tt.codeTheme {
				t.Errorf("NewConverter() codeTheme = %q, want %q", conv.codeTheme, tt.codeTheme)
			}
			// Should use default light theme
			if conv.theme != LightTheme {
				t.Error("NewConverter() should use light theme by default")
			}
		})
	}
}

func TestNewConverterWithAllOptions(t *testing.T) {
	tests := []struct {
		name        string
		codeTheme   string
		themeName   string
		expectTheme Theme
	}{
		{"light theme", "monokai", "light", LightTheme},
		{"dark theme", "github", "dark", DarkTheme},
		{"unknown theme defaults to light", "monokai", "unknown", LightTheme},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConverter(WithCodeTheme(tt.codeTheme), WithTheme(tt.themeName))
			if conv == nil {
				t.Error("NewConverter() returned nil")
			}
			if conv.codeTheme != tt.codeTheme {
				t.Errorf("NewConverter() codeTheme = %q, want %q", conv.codeTheme, tt.codeTheme)
			}
			if conv.theme != tt.expectTheme {
				t.Errorf("NewConverter() theme mismatch")
			}
		})
	}
}

func TestGetAvailableThemes(t *testing.T) {
	themes := GetAvailableThemes()
	if len(themes) == 0 {
		t.Error("GetAvailableThemes() returned empty list")
	}

	// Check that light and dark themes are available
	hasLight := false
	hasDark := false
	for _, theme := range themes {
		if theme == "light" {
			hasLight = true
		}
		if theme == "dark" {
			hasDark = true
		}
	}

	if !hasLight {
		t.Error("GetAvailableThemes() should include 'light' theme")
	}
	if !hasDark {
		t.Error("GetAvailableThemes() should include 'dark' theme")
	}
}

func TestRenderCodeWithDifferentStyles(t *testing.T) {
	// Test rendering with different styles
	slideContent := `# Style Test
Test
16 Feb 2026

Author

## Go Code

	package main
	
	import "fmt"
	
	func main() {
		fmt.Println("Hello")
	}
`

	styles := []string{"monokai", "github", "dracula"}

	for _, codeTheme := range styles {
		t.Run(codeTheme, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "style-*.slide")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tmpFile.Close()

			outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
			defer os.Remove(outputPath)

			conv := NewConverter(WithCodeTheme(codeTheme))
			err = conv.Convert(tmpFile.Name(), outputPath)
			if err != nil {
				t.Errorf("Convert() failed for style %s: %v", codeTheme, err)
			}

			// Check if output file exists
			info, err := os.Stat(outputPath)
			if os.IsNotExist(err) {
				t.Errorf("Output PDF file was not created for style %s", codeTheme)
			}
			if info.Size() < 1024 {
				t.Errorf("PDF file too small for style %s: %d bytes", codeTheme, info.Size())
			}
		})
	}
}

func TestParseHTMLFormatting(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantFrags []TextFragment
	}{
		{
			name:  "plain text",
			input: "Hello world",
			wantFrags: []TextFragment{
				{Text: "Hello world"},
			},
		},
		{
			name:  "bold text",
			input: "<strong>bold</strong>",
			wantFrags: []TextFragment{
				{Text: "bold", Bold: true},
			},
		},
		{
			name:  "italic text",
			input: "<em>italic</em>",
			wantFrags: []TextFragment{
				{Text: "italic", Italic: true},
			},
		},
		{
			name:  "link with href",
			input: `<a href="https://example.com">click here</a>`,
			wantFrags: []TextFragment{
				{Text: "click here", URL: "https://example.com"},
			},
		},
		{
			name:  "link with single-quoted href",
			input: `<a href='https://golang.org'>Go</a>`,
			wantFrags: []TextFragment{
				{Text: "Go", URL: "https://golang.org"},
			},
		},
		{
			name:  "text before and after link",
			input: `Visit <a href="https://go.dev">Go</a> now.`,
			wantFrags: []TextFragment{
				{Text: "Visit "},
				{Text: "Go", URL: "https://go.dev"},
				{Text: " now."},
			},
		},
		{
			name:  "bold link",
			input: `<strong><a href="https://example.com">bold link</a></strong>`,
			wantFrags: []TextFragment{
				{Text: "bold link", Bold: true, URL: "https://example.com"},
			},
		},
		{
			name:  "multiple links",
			input: `<a href="https://a.com">A</a> and <a href="https://b.com">B</a>`,
			wantFrags: []TextFragment{
				{Text: "A", URL: "https://a.com"},
				{Text: " and "},
				{Text: "B", URL: "https://b.com"},
			},
		},
		{
			name:  "html entities decoded",
			input: `x &lt; 10 &amp;&amp; y &gt; 5`,
			wantFrags: []TextFragment{
				{Text: "x < 10 && y > 5"},
			},
		},
		{
			name:  "inline code",
			input: "<code>(r Rectangle)</code>",
			wantFrags: []TextFragment{
				{Text: "(r Rectangle)", Code: true},
			},
		},
		{
			name:  "inline code inside text",
			input: "call <code>Foo()</code> here",
			wantFrags: []TextFragment{
				{Text: "call "},
				{Text: "Foo()", Code: true},
				{Text: " here"},
			},
		},
		{
			name:      "empty input",
			input:     "",
			wantFrags: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHTMLFormatting(tt.input)

			if len(got) != len(tt.wantFrags) {
				t.Fatalf("parseHTMLFormatting() returned %d fragments, want %d\ngot:  %+v\nwant: %+v",
					len(got), len(tt.wantFrags), got, tt.wantFrags)
			}

			for i, frag := range got {
				want := tt.wantFrags[i]
				if frag.Text != want.Text {
					t.Errorf("fragment[%d].Text = %q, want %q", i, frag.Text, want.Text)
				}
				if frag.Bold != want.Bold {
					t.Errorf("fragment[%d].Bold = %v, want %v", i, frag.Bold, want.Bold)
				}
				if frag.Italic != want.Italic {
					t.Errorf("fragment[%d].Italic = %v, want %v", i, frag.Italic, want.Italic)
				}
				if frag.Code != want.Code {
					t.Errorf("fragment[%d].Code = %v, want %v", i, frag.Code, want.Code)
				}
				if frag.URL != want.URL {
					t.Errorf("fragment[%d].URL = %q, want %q", i, frag.URL, want.URL)
				}
			}
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain text", "hello world", "hello world"},
		{"bold tag", "<b>bold</b>", "bold"},
		{"anchor tag", `<a href="https://example.com">link text</a>`, "link text"},
		{"nested tags", "<p><strong>bold</strong> and <em>italic</em></p>", "bold and italic"},
		{"html entities", "&lt;foo&gt; &amp; &quot;bar&quot;", "<foo> & \"bar\""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHTMLTags(tt.input)
			if got != tt.expected {
				t.Errorf("stripHTMLTags(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestThemeLinkColors(t *testing.T) {
	t.Run("light theme has link color", func(t *testing.T) {
		c := LightTheme.LinkColor
		if c.R == 0 && c.G == 0 && c.B == 0 {
			t.Error("LightTheme.LinkColor is black (not set)")
		}
		if c.R > 255 || c.G > 255 || c.B > 255 {
			t.Errorf("LightTheme.LinkColor out of range: %+v", c)
		}
	})

	t.Run("dark theme has link color", func(t *testing.T) {
		c := DarkTheme.LinkColor
		if c.R == 0 && c.G == 0 && c.B == 0 {
			t.Error("DarkTheme.LinkColor is black (not set)")
		}
		if c.R > 255 || c.G > 255 || c.B > 255 {
			t.Errorf("DarkTheme.LinkColor out of range: %+v", c)
		}
	})

	t.Run("light and dark link colors differ", func(t *testing.T) {
		l := LightTheme.LinkColor
		d := DarkTheme.LinkColor
		if l == d {
			t.Error("LightTheme and DarkTheme have identical link colors")
		}
	})
}

func TestRenderLinkDirective(t *testing.T) {
	// .link directive is the legacy-format way to add hyperlinks
	slideContent := `Legacy Presentation
Links test
16 Feb 2026

Author

* Links Slide

.link https://golang.org The Go Programming Language
.link https://github.com GitHub
`

	tmpFile, err := os.CreateTemp("", "link-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Fatalf("Convert() failed for .link directive: %v", err)
	}

	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Fatal("Output PDF file was not created")
	}
	if info.Size() < 1024 {
		t.Errorf("PDF file too small: %d bytes", info.Size())
	}
}

func TestRenderLinkDirectiveWithoutLabel(t *testing.T) {
	// .link without a label — URL itself should be used as display text
	slideContent := `Legacy Presentation
Links test
16 Feb 2026

Author

* Links Slide

.link https://golang.org
`

	tmpFile, err := os.CreateTemp("", "link-nolabel-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Fatalf("Convert() failed for .link without label: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output PDF file was not created")
	}
}

func TestMarkdownLinksInParagraph(t *testing.T) {
	// Markdown links [text](url) should produce clickable links in the PDF
	slideContent := `# Markdown Links
16 Feb 2026

Author

## Links Slide

Visit [Go](https://golang.org) and [GitHub](https://github.com) for more.

Plain text after the links.
`

	tmpFile, err := os.CreateTemp("", "mdlinks-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Fatalf("Convert() failed for markdown links: %v", err)
	}

	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Fatal("Output PDF file was not created")
	}
	if info.Size() < 1024 {
		t.Errorf("PDF file too small: %d bytes", info.Size())
	}
}

func TestMarkdownLinksInList(t *testing.T) {
	// Links inside list items should also work
	slideContent := `# Markdown Links in Lists
16 Feb 2026

Author

## Resources

- [Go Documentation](https://pkg.go.dev)
- [GitHub](https://github.com)
- Plain item without link
`

	tmpFile, err := os.CreateTemp("", "mdlinks-list-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Fatalf("Convert() failed for markdown links in list: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output PDF file was not created")
	}
}

func TestRenderLinkUnit(t *testing.T) {
	// Unit test calling renderLink directly
	conv := NewConverter()
	conv.pdf = gofpdf.New("L", "mm", "A4", "")
	conv.pdf.AddPage()
	conv.translator = conv.pdf.UnicodeTranslatorFromDescriptor("cp1251")
	conv.pdf.SetFont("Helvetica", "", 18)

	tests := []struct {
		name    string
		label   string
		rawURL  string
		startY  float64
		wantAdv bool // y should advance
	}{
		{"with label", "Click here", "https://example.com", 50.0, true},
		{"without label (URL as display)", "", "https://example.com", 70.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := present.Link{Label: tt.label}
			if tt.rawURL != "" {
				u, err := url.Parse(tt.rawURL)
				if err != nil {
					t.Fatalf("failed to parse URL %q: %v", tt.rawURL, err)
				}
				link.URL = u
			}

			newY := conv.renderLink(link, tt.startY)
			if tt.wantAdv && newY <= tt.startY {
				t.Errorf("renderLink() did not advance Y: got %.1f, started at %.1f", newY, tt.startY)
			}
		})
	}
}

func TestRenderHTMLBlockquote(t *testing.T) {
	conv := NewConverter()
	conv.pdf = gofpdf.New("L", "mm", "A4", "")
	conv.pdf.AddPage()
	conv.translator = conv.pdf.UnicodeTranslatorFromDescriptor("")

	tests := []struct {
		name string
		html string
	}{
		{
			name: "simple blockquote with paragraph",
			html: "<blockquote>\n<p>This is a quoted text block.</p>\n</blockquote>",
		},
		{
			name: "blockquote with multiple paragraphs",
			html: "<blockquote>\n<p>First paragraph of the quote.</p>\n<p>Second paragraph of the quote.</p>\n</blockquote>",
		},
		{
			name: "blockquote with bold text",
			html: "<blockquote>\n<p>Quote with <strong>bold</strong> content.</p>\n</blockquote>",
		},
		{
			name: "blockquote with no p tags",
			html: "<blockquote>Plain text without paragraph tags.</blockquote>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startY := 45.0
			endY := conv.renderHTMLBlockquote(tt.html, startY)
			if endY <= startY {
				t.Errorf("renderHTMLBlockquote() did not advance Y: startY=%v, endY=%v", startY, endY)
			}
		})
	}
}

func TestConvertMarkdownInlineCode(t *testing.T) {
	slideContent := `# Inline Code Test
Test Presentation
18 Feb 2026

Author Name

## Slide With Inline Code

Use ` + "`(r Rectangle)`" + ` as a receiver.

- Call ` + "`Foo()`" + ` to start
- Set ` + "`x := 42`" + ` before calling

Mixed: use **bold** and ` + "`code`" + ` together.
`

	tmpFile, err := os.CreateTemp("", "inlinecode-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() error = %v", err)
	}

	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created")
	} else if info.Size() < 1024 {
		t.Errorf("PDF file too small: %d bytes", info.Size())
	}
}

func TestConvertMarkdownBlockquote(t *testing.T) {
	slideContent := `# Blockquote Test
Test Presentation
18 Feb 2026

Author Name
author@example.com

## Slide With Blockquote

Regular paragraph before quote.

> This is a blockquote block.
> It can span multiple lines.

Regular paragraph after quote.

## Slide With Bold Blockquote

> **Important:** This is a bold quote.
`

	tmpFile, err := os.CreateTemp("", "blockquote-*.slide")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	err = conv.Convert(tmpFile.Name(), outputPath)
	if err != nil {
		t.Errorf("Convert() error = %v", err)
	}

	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Errorf("Output PDF file was not created")
	} else if info.Size() < 1024 {
		t.Errorf("PDF file too small: %d bytes", info.Size())
	}
}

// createTestPNG creates a small solid-color PNG image at the given path.
func createTestPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: 100, G: 149, B: 237, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("createTestPNG: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("createTestPNG encode: %v", err)
	}
}

func TestRenderImageLegacy(t *testing.T) {
	// Create temp dir with a test image and slide
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	createTestPNG(t, imgPath, 400, 300)

	slideContent := "Legacy With Image\nTest\n18 Feb 2026\n\nAuthor\n\n* Image Slide\n\n.image test.png\n"
	slideFile := filepath.Join(dir, "test.slide")
	if err := os.WriteFile(slideFile, []byte(slideContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	outputPath := filepath.Join(dir, "out.pdf")
	conv := NewConverter()
	if err := conv.Convert(slideFile, outputPath); err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Fatal("Output PDF not created")
	}
	if info.Size() < 1024 {
		t.Errorf("PDF too small: %d bytes", info.Size())
	}
}

func TestRenderImageMarkdown(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "photo.png")
	createTestPNG(t, imgPath, 800, 600)

	slideContent := "# Markdown With Image\n18 Feb 2026\n\nAuthor\n\n## Slide With Image\n\n![alt text](photo.png)\n"
	slideFile := filepath.Join(dir, "test.slide")
	if err := os.WriteFile(slideFile, []byte(slideContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	outputPath := filepath.Join(dir, "out.pdf")
	conv := NewConverter()
	if err := conv.Convert(slideFile, outputPath); err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Fatal("Output PDF not created")
	}
	if info.Size() < 1024 {
		t.Errorf("PDF too small: %d bytes", info.Size())
	}
}

func TestRenderImageMissingFile(t *testing.T) {
	// Missing image should log warning but not fail conversion
	slideContent := "Legacy Presentation\nTest\n18 Feb 2026\n\nAuthor\n\n* Slide\n\n.image nonexistent.png\n"
	tmpFile, err := os.CreateTemp("", "imgmissing-*.slide")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter(WithQuiet(true))
	if err := conv.Convert(tmpFile.Name(), outputPath); err != nil {
		t.Errorf("Convert() should not fail for missing image, got: %v", err)
	}
}

func TestRenderHTMLImage(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "icon.png")
	createTestPNG(t, imgPath, 100, 100)

	conv := NewConverter()
	conv.slideDir = dir

	cleanup, err := conv.initPDF()
	if err != nil {
		t.Fatalf("initPDF: %v", err)
	}
	defer cleanup()
	conv.pdf.AddPage()

	imgHTML := `<img src="icon.png" alt="icon">`
	newY := conv.renderHTMLImage(imgHTML, 50.0)
	if newY <= 50.0 {
		t.Errorf("renderHTMLImage() did not advance Y: got %.1f, started at 50.0", newY)
	}
}

// --------------------------------------------------------------------------
// Tests for preprocessMarkdownComments (bug-fix: // in ``` blocks stripped)
// --------------------------------------------------------------------------

func TestPreprocessMarkdownComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "legacy format - not modified",
			input: "My Presentation\n```go\n// comment\npackage main\n```\n",
			want:  "My Presentation\n```go\n// comment\npackage main\n```\n",
		},
		{
			name:  "markdown - // outside code block not escaped",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n// This is a slide comment\n",
			want:  "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n// This is a slide comment\n",
		},
		{
			name:  "markdown - // inside code block escaped",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```go\n// comment\npackage main\n```\n",
			want:  "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```go\n\u200C// comment\npackage main\n```\n",
		},
		{
			name: "markdown - multiple // lines in one block",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```go\n" +
				"// Package config\n// provides config.\npackage config\n```\n",
			want: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```go\n" +
				"\u200C// Package config\n\u200C// provides config.\npackage config\n```\n",
		},
		{
			name: "markdown - // in multiple separate code blocks",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide 1\n\n```go\n// c1\n```\n\n## Slide 2\n\n```go\n// c2\n```\n",
			want: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide 1\n\n```go\n\u200C// c1\n```\n\n## Slide 2\n\n```go\n\u200C// c2\n```\n",
		},
		{
			name: "markdown - slide comment outside vs code comment inside",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n// slide comment\n```go\n// code comment\n```\n",
			want: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n// slide comment\n```go\n\u200C// code comment\n```\n",
		},
		{
			name:  "markdown - non-// comment (/*) not escaped",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```go\n/* block comment */\npackage main\n```\n",
			want:  "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```go\n/* block comment */\npackage main\n```\n",
		},
		{
			name:  "markdown - language specifier in opening fence",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```python\n// not python but has slashes\n```\n",
			want:  "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```python\n\u200C// not python but has slashes\n```\n",
		},
		{
			name:  "empty content",
			input: "",
			want:  "",
		},
		{
			name:  "markdown - file path comment (original bug case)",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```go\n// ./my-module/internal/config/file.go\npackage config\n```\n",
			want:  "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```go\n\u200C// ./my-module/internal/config/file.go\npackage config\n```\n",
		},
		{
			name:  "markdown - # comment inside bash code block escaped",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```bash\n# shell comment\necho hello\n```\n",
			want:  "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```bash\n\u200C# shell comment\necho hello\n```\n",
		},
		{
			name:  "markdown - shebang inside code block escaped",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```bash\n#!/usr/bin/env bash\necho hi\n```\n",
			want:  "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```bash\n\u200C#!/usr/bin/env bash\necho hi\n```\n",
		},
		{
			name:  "markdown - ## inside code block escaped",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```c\n## pragma-style comment\nint x;\n```\n",
			want:  "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```c\n\u200C## pragma-style comment\nint x;\n```\n",
		},
		{
			name:  "markdown - # outside code block not escaped",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\nSome text.\n",
			want:  "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\nSome text.\n",
		},
		{
			name: "markdown - # and $ in bash block: hash escaped, dollar kept",
			input: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```bash\n# Создание нового модуля\n$ go mod init github.com/username/project-name\n```\n",
			want: "# Title\n15 Feb 2026\n\nAuthor\n\n## Slide\n\n```bash\n\u200C# Создание нового модуля\n$ go mod init github.com/username/project-name\n```\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(preprocessMarkdownComments([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("preprocessMarkdownComments():\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestRenderHTMLCodeStripsEscapePrefix(t *testing.T) {
	// renderHTMLCode must strip the \u200C prefix inserted by
	// preprocessMarkdownComments so it does not appear in the rendered PDF.
	conv := NewConverter()
	cleanup, err := conv.initPDF()
	if err != nil {
		t.Fatalf("initPDF: %v", err)
	}
	defer cleanup()
	conv.pdf.AddPage()

	html := "<pre><code class=\"language-go\">" +
		"\u200C// ./my-module/internal/config/file.go\npackage config\n" +
		"</code></pre>"

	startY := 45.0
	newY := conv.renderHTMLCode(html, startY)
	if newY <= startY {
		t.Errorf("renderHTMLCode() did not advance Y: got %.1f, started at %.1f", newY, startY)
	}
}

func TestConvertMarkdownCodeBlockWithGoComments(t *testing.T) {
	// Regression test: // lines inside ``` code blocks must survive
	// the present parser and appear in the generated PDF without error.
	slideContent := "# Config Module\n" +
		"15 Feb 2026\n\n" +
		"Author\n\n" +
		"## File Header Comment\n\n" +
		"```go\n" +
		"// ./my-module/internal/config/file.go\n" +
		"package config\n" +
		"```\n\n" +
		"## Multiple Comments\n\n" +
		"```go\n" +
		"// Package config provides configuration management.\n" +
		"// It reads from environment variables and config files.\n" +
		"package config\n\n" +
		"import \"os\"\n\n" +
		"// Config holds the application configuration.\n" +
		"type Config struct {\n" +
		"    Port string\n" +
		"}\n" +
		"```\n"

	tmpFile, err := os.CreateTemp("", "gocomments-*.slide")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	if err := conv.Convert(tmpFile.Name(), outputPath); err != nil {
		t.Fatalf("Convert() failed for slide with // code comments: %v", err)
	}

	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Fatal("Output PDF file was not created")
	}
	if info.Size() < 1024 {
		t.Errorf("PDF file too small: %d bytes", info.Size())
	}
}

func TestPreprocessMarkdownCommentsDoesNotAffectPresentParsing(t *testing.T) {
	// End-to-end check: after preprocessing, present.Parse must include the
	// // comment line in the generated HTML (it must NOT be stripped).
	slideContent := "# Title\n" +
		"15 Feb 2026\n\n" +
		"Author\n\n" +
		"## Slide\n\n" +
		"```go\n" +
		"// ./my-module/internal/config/file.go\n" +
		"package config\n" +
		"```\n"

	preprocessed := preprocessMarkdownComments([]byte(slideContent))

	ctx := present.Context{
		ReadFile: func(name string) ([]byte, error) {
			return os.ReadFile(name)
		},
	}
	doc, err := ctx.Parse(bytes.NewReader(preprocessed), "test.slide", 0)
	if err != nil {
		t.Fatalf("present.Parse() error: %v", err)
	}

	if len(doc.Sections) == 0 {
		t.Fatal("present.Parse() returned no sections")
	}

	// Walk all elements looking for an HTML element that contains the comment.
	found := false
	for _, section := range doc.Sections {
		for _, elem := range section.Elem {
			if h, ok := elem.(present.HTML); ok {
				html := string(h.HTML)
				// The \u200C prefix must be present at this stage (it's stripped
				// later by renderHTMLCode, not by present).
				if strings.Contains(html, "\u200C// ./my-module/internal/config/file.go") {
					found = true
				}
			}
		}
	}

	if !found {
		t.Error("present.Parse() stripped the // comment line from the code block; " +
			"preprocessMarkdownComments did not protect it correctly")
	}
}

func TestConvertMarkdownBashCodeBlockWithHashLines(t *testing.T) {
	// Regression test: lines starting with # or #! inside ``` bash code blocks
	// must not be stripped/broken by the present parser.
	slideContent := "# Go Module Setup\n" +
		"15 Feb 2026\n\n" +
		"Author\n\n" +
		"## Создание модуля\n\n" +
		"```bash\n" +
		"# Создание нового модуля\n" +
		"$ go mod init github.com/username/project-name\n" +
		"```\n\n" +
		"## Shebang Example\n\n" +
		"```bash\n" +
		"#!/usr/bin/env bash\n" +
		"echo \"Hello\"\n" +
		"```\n"

	tmpFile, err := os.CreateTemp("", "bashcode-*.slide")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	tmpFile.Close()

	outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
	defer os.Remove(outputPath)

	conv := NewConverter()
	if err := conv.Convert(tmpFile.Name(), outputPath); err != nil {
		t.Fatalf("Convert() failed for bash code block with # lines: %v", err)
	}

	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Fatal("Output PDF file was not created")
	}
	if info.Size() < 1024 {
		t.Errorf("PDF file too small: %d bytes", info.Size())
	}
}

func TestBashHashLinesPreservedInPresentHTML(t *testing.T) {
	// Check that after preprocessing, present.Parse keeps the # lines inside
	// the code block HTML (they must NOT be treated as headings).
	slideContent := "# Title\n" +
		"15 Feb 2026\n\n" +
		"Author\n\n" +
		"## Slide\n\n" +
		"```bash\n" +
		"# Создание нового модуля\n" +
		"$ go mod init github.com/username/project-name\n" +
		"```\n"

	preprocessed := preprocessMarkdownComments([]byte(slideContent))

	ctx := present.Context{
		ReadFile: func(name string) ([]byte, error) {
			return os.ReadFile(name)
		},
	}
	doc, err := ctx.Parse(bytes.NewReader(preprocessed), "test.slide", 0)
	if err != nil {
		t.Fatalf("present.Parse() error: %v", err)
	}

	if len(doc.Sections) == 0 {
		t.Fatal("present.Parse() returned no sections")
	}

	// There must be exactly one HTML element whose content is a non-empty
	// <pre><code> block that contains both the # line and the $ line.
	found := false
	for _, section := range doc.Sections {
		for _, elem := range section.Elem {
			if h, ok := elem.(present.HTML); ok {
				html := string(h.HTML)
				if strings.Contains(html, "<pre><code") &&
					strings.Contains(html, "\u200C# Создание нового модуля") &&
					strings.Contains(html, "$ go mod init github.com/username/project-name") {
					found = true
				}
			}
		}
	}

	if !found {
		t.Error("present.Parse() did not preserve the # line inside the bash code block; " +
			"preprocessMarkdownComments did not protect it correctly")
	}
}

func TestRenderWithDifferentPDFThemes(t *testing.T) {
	// Test rendering with different PDF themes
	slideContent := `# Theme Test
Test Presentation
16 Feb 2026

## Introduction

This is a test presentation to verify PDF themes.

## Code Example

Example code:

	package main

	func main() {
		fmt.Println("Hello, World!")
	}

## Conclusion

Different themes should have different colors.
`

	themes := []string{"light", "dark"}

	for _, themeName := range themes {
		t.Run(themeName, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "theme-*.slide")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write([]byte(slideContent)); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tmpFile.Close()

			outputPath := strings.TrimSuffix(tmpFile.Name(), ".slide") + ".pdf"
			defer os.Remove(outputPath)

			conv := NewConverter(WithCodeTheme("monokai"), WithTheme(themeName))
			err = conv.Convert(tmpFile.Name(), outputPath)
			if err != nil {
				t.Errorf("Convert() failed for theme %s: %v", themeName, err)
			}

			// Check if output file exists
			info, err := os.Stat(outputPath)
			if os.IsNotExist(err) {
				t.Errorf("Output PDF file was not created for theme %s", themeName)
			} else if info.Size() < 1024 {
				t.Errorf("PDF file too small for theme %s: %d bytes", themeName, info.Size())
			}
		})
	}
}
