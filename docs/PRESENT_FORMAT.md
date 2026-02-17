# Present Format Guide

This document explains the difference between Legacy and Markdown-enabled present formats.

## Format Detection

The format is determined by the **first line** of the file:

- **Markdown-enabled**: Title starts with `# `
- **Legacy**: Title without `# ` prefix

## Comparison

### File Header

**Markdown-enabled:**
```
# Title of Presentation
Subtitle
15 Feb 2026

Author Name
email@example.com
```

**Legacy:**
```
Title of Presentation
Subtitle
15 Feb 2026

Author Name
email@example.com
```

### Slide/Section Headers

**Markdown-enabled:**
- Main sections: `##`
- Subsections: `###`
- Sub-subsections: `####`

```markdown
## Main Section

### Subsection

#### Sub-subsection
```

**Legacy:**
- Main sections: `*`
- Subsections: `**`
- Sub-subsections: `***`

```
* Main Section

** Subsection

*** Sub-subsection
```

### Text Formatting

**Markdown-enabled (CommonMark):**
- Italic: `_text_`
- Bold: `**text**`
- Inline code: `` `code` ``
- Links: `[label](url)`

**Legacy:**
- Italic: `_text_`
- Bold: `*text*`
- Inline code: `` `code` ``
- Links: `[[url][label]]` or `[[url]]`

### Comments

**Markdown-enabled:**
```markdown
// This is a comment
// It will be completely ignored
```

**Legacy:**
```
# This is a comment
# It will be ignored
```

### Anchor IDs

**Markdown-enabled** supports custom anchor IDs:
```markdown
## Section Title {#custom-id}
```

**Legacy** does not support custom anchors.

### Lists

Both formats support bulleted lists with `-`:

```
- First item
- Second item
- Third item continued
  on the next line
```

### Code Blocks

Both formats use indentation (tabs or 4+ spaces):

```
	package main
	
	func main() {
		fmt.Println("Hello")
	}
```

### Present Commands

Both formats support the same commands:

```
.code file.go
.play example.go
.image picture.jpg
.link https://example.com
.caption Text here
```

### Speaker Notes

Both formats use `: ` prefix:

```
: These are speaker notes
: Visible only in presenter view
```

## Examples in This Project

The example file in `example/` directory uses **Markdown-enabled** format:

- `example/presentation.slide` - Comprehensive Markdown features demonstration

## Recommendations

✅ **Use Markdown-enabled format** for new presentations:
- Modern CommonMark syntax
- Better compatibility with Markdown tools
- Support for custom anchor IDs
- Cleaner comment syntax (`//` instead of `#`)
- More explicit section hierarchy (`##`, `###`)

⚠️ **Legacy format** is supported but not recommended for new files.

## Migration

To migrate from legacy to Markdown-enabled:

1. Add `# ` prefix to the title (first line)
2. Replace `*` with `##` for sections
3. Replace `**` with `###` for subsections
4. Replace `#` comments with `//`
5. Optional: Convert `*bold*` to `**bold**`
6. Optional: Convert `[[url][label]]` to `[label](url)`

## References

- [Present Package Documentation](https://pkg.go.dev/golang.org/x/tools/present)
- [CommonMark Specification](https://commonmark.org/)
- [CommonMark Tutorial](https://commonmark.org/help/tutorial/)
