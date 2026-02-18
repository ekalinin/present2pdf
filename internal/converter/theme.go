package converter

import "github.com/alecthomas/chroma/v2/styles"

// RGB represents an RGB color
type RGB struct {
	R, G, B int
}

// Theme defines the color scheme for PDF presentation
type Theme struct {
	// Title slide colors
	TitleBackground RGB
	TitleText       RGB
	TitleSubtext    RGB
	TitleDate       RGB

	// Content slide colors
	SlideBackground RGB
	SlideTitle      RGB
	SlideTitleLine  RGB
	SlideText       RGB

	// Code block colors
	CodeBackground RGB
	CodeText       RGB
	CodeLineNumber RGB

	// Link color
	LinkColor RGB

	// Blockquote colors
	BlockquoteBackground RGB
	BlockquoteBorder     RGB
}

// Predefined themes
var (
	// LightTheme is the default light theme
	LightTheme = Theme{
		TitleBackground: RGB{41, 128, 185},  // Blue
		TitleText:       RGB{255, 255, 255}, // White
		TitleSubtext:    RGB{255, 255, 255}, // White
		TitleDate:       RGB{255, 255, 255}, // White
		SlideBackground: RGB{255, 255, 255}, // White
		SlideTitle:      RGB{41, 128, 185},  // Blue
		SlideTitleLine:  RGB{41, 128, 185},  // Blue
		SlideText:       RGB{0, 0, 0},       // Black
		CodeBackground:       RGB{40, 44, 52},    // Dark gray
		CodeText:             RGB{171, 178, 191}, // Light gray
		CodeLineNumber:       RGB{128, 128, 128}, // Gray
		LinkColor:            RGB{0, 102, 204},   // Link blue
		BlockquoteBackground: RGB{240, 247, 255}, // Light blue-white
		BlockquoteBorder:     RGB{41, 128, 185},  // Blue (same as title)
	}

	// DarkTheme is a dark theme
	DarkTheme = Theme{
		TitleBackground: RGB{30, 30, 46},    // Dark blue-gray
		TitleText:       RGB{205, 214, 244}, // Light gray
		TitleSubtext:    RGB{166, 173, 200}, // Medium gray
		TitleDate:       RGB{137, 180, 250}, // Light blue
		SlideBackground: RGB{36, 39, 58},    // Dark gray-blue
		SlideTitle:      RGB{137, 180, 250}, // Light blue
		SlideTitleLine:  RGB{137, 180, 250}, // Light blue
		SlideText:       RGB{205, 214, 244}, // Light gray
		CodeBackground:       RGB{30, 30, 46},    // Darker blue-gray
		CodeText:             RGB{205, 214, 244}, // Light gray
		CodeLineNumber:       RGB{108, 112, 134}, // Medium gray
		LinkColor:            RGB{137, 180, 250}, // Light blue
		BlockquoteBackground: RGB{48, 52, 72},    // Slightly lighter than slide bg
		BlockquoteBorder:     RGB{137, 180, 250}, // Light blue (same as title)
	}

	// availableThemes maps theme names to themes
	availableThemes = map[string]Theme{
		"light": LightTheme,
		"dark":  DarkTheme,
	}
)

// GetAvailableStyles returns a list of available syntax highlighting styles
func GetAvailableStyles() []string {
	return styles.Names()
}

// GetAvailableThemes returns a list of available PDF themes
func GetAvailableThemes() []string {
	themes := make([]string, 0, len(availableThemes))
	for name := range availableThemes {
		themes = append(themes, name)
	}
	return themes
}
