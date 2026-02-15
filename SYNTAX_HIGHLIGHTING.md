# Syntax Highlighting

present2pdf now supports syntax highlighting for code blocks in your presentations!

## How it works

The tool automatically highlights code with proper syntax colors using the Chroma library. Code blocks are rendered with:

- **Dark background** (similar to Monokai/VS Code dark theme)
- **Colored syntax highlighting** for keywords, strings, comments, functions, etc.
- **Automatic fallback** to plain rendering if highlighting fails

## Supported Languages

The tool supports syntax highlighting for many programming languages:

- Go (default)
- Python
- JavaScript / TypeScript
- Java
- C / C++
- Rust
- Ruby
- PHP
- Bash / Shell
- HTML / CSS
- JSON / XML / YAML
- SQL
- And many more...

## Language Detection

The language is detected automatically from:

1. **File extension** in code blocks (if specified)
2. **Class attribute** in HTML code blocks (e.g., `class="language-python"`)
3. **Default to Go** if no language information is available

## Color Scheme

The default color scheme is based on Monokai with a dark background:

- **Background**: Dark gray (#282C34)
- **Keywords**: Purple (#C678DD)
- **Strings**: Green (#98C379)
- **Comments**: Gray (#5C636F)
- **Functions**: Blue (#61AFEF)
- **Numbers**: Orange (#D19A66)
- **Built-ins**: Yellow (#E5C07B)
- **Default text**: Light gray (#ABB2BF)

## Examples

### Go Code

```
## Go Example

	package main

	import "fmt"

	func greet(name string) string {
		return fmt.Sprintf("Hello, %s!", name)
	}

	func main() {
		fmt.Println(greet("World"))
	}
```

### Python Code

```
## Python Example

	def fibonacci(n):
		"""Calculate fibonacci numbers"""
		if n <= 1:
			return n
		return fibonacci(n-1) + fibonacci(n-2)

	for i in range(10):
		print(f"fib({i}) = {fibonacci(i)}")
```

### JavaScript Code

```
## JavaScript Example

	const fetchData = async (url) => {
		try {
			const response = await fetch(url);
			return await response.json();
		} catch (error) {
			console.error('Error:', error);
			throw error;
		}
	};
```

## Technical Details

The syntax highlighting is implemented using:

- **Chroma v2** - A general-purpose syntax highlighter written in Go
- **Lexer-based tokenization** - Precise token identification
- **PDF text positioning** - Each token is rendered with its specific color
- **Efficient caching** - Tokens are processed once per code block

## Limitations

- Maximum 12 lines of code per block (with "..." indicator for overflow)
- Line wrapping is not supported (long lines may be truncated)
- Some advanced formatting features (e.g., background highlights) are not supported

## Future Improvements

Planned enhancements:

- Support for more color schemes (light themes, etc.)
- Line numbers in code blocks
- Better handling of long lines (wrapping or horizontal scrolling)
- Customizable syntax highlighting colors
- Support for code annotations and highlights

