package converter

import (
	"os"
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
