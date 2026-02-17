# Syntax Highlighting

present2pdf supports advanced syntax highlighting for code blocks in your presentations!

## How it works

The tool automatically highlights code with proper syntax colors using the Chroma library. Code blocks are rendered with:

- **Customizable color schemes** - Choose from 70+ built-in styles
- **Colored syntax highlighting** for keywords, strings, comments, functions, etc.
- **Automatic fallback** to plain rendering if highlighting fails

## Choosing a Color Scheme

### Command Line

Use the `-code-theme` flag to select a color scheme:

```bash
# Use Monokai (default)
./present2pdf -input presentation.slide

# Use GitHub theme (light background)
./present2pdf -input presentation.slide -code-theme github

# Use Dracula theme
./present2pdf -input presentation.slide -code-theme dracula

# List all available themes
./present2pdf -list-code-themes
```

### Available Themes

Run `./present2pdf -list-code-themes` to see all 70+ options. Popular choices include:

#### Dark Themes
- `monokai` (default) - Vibrant colors on dark background
- `dracula` - Purple and cyan on dark background
- `github-dark` - GitHub's dark theme
- `nord` - Arctic-inspired blue tones
- `tokyonight-night` - Dark blue with bright accents
- `doom-one` - Emacs Doom theme
- `gruvbox` - Retro groove colors
- `solarized-dark` - Popular dark variant

#### Light Themes
- `github` - GitHub's light theme
- `solarized-light` - Popular light variant
- `xcode` - Apple Xcode style
- `emacs` - Classic Emacs colors
- `colorful` - Bright and vibrant

#### Classic Themes
- `vim` - Classic Vim colors
- `tango` - GNOME Tango palette
- `friendly` - Easy on the eyes
- `pastie` - Pastie clone

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

## Color Scheme Examples

### Monokai (Default)

The default color scheme based on Monokai:

- **Background**: Dark gray (#282C34)
- **Keywords**: Purple/Magenta
- **Strings**: Green/Yellow
- **Comments**: Gray
- **Functions**: Blue/Cyan
- **Numbers**: Orange
- **Built-ins**: Yellow
- **Default text**: Light gray

### GitHub

Light theme matching GitHub:

- **Background**: White
- **Keywords**: Purple
- **Strings**: Blue
- **Comments**: Gray
- **Functions**: Purple/Blue
- **Numbers**: Teal

### Dracula

Dark theme with purple and cyan:

- **Background**: Dark (#282A36)
- **Keywords**: Pink
- **Strings**: Yellow
- **Comments**: Gray/Blue
- **Functions**: Green
- **Numbers**: Purple

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

- Maximum 20 lines of code per block (with "..." indicator for overflow)
- Line wrapping is not supported (long lines may be truncated)
- Some advanced formatting features (e.g., background highlights) are not supported

## Future Improvements

Planned enhancements:

- Support for more color schemes (light themes, etc.)
- Line numbers in code blocks
- Better handling of long lines (wrapping or horizontal scrolling)
- Customizable syntax highlighting colors
- Support for code annotations and highlights

