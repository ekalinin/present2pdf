package converter

import (
	"os"
	"strings"
	"testing"

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
