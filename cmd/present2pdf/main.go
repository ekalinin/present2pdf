package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ekalinin/present2pdf/internal/converter"
)

// Version is set during build time via ldflags
var version = "dev"

func main() {
	inputFile := flag.String("input", "", "Path to .slide file (required)")
	outputFile := flag.String("output", "", "Path to output PDF file (optional, defaults to input filename with .pdf extension)")
	codeTheme := flag.String("code-theme", "monokai", "Code syntax highlighting theme (use -list-code-themes to see available options)")
	pdfTheme := flag.String("theme", "light", "PDF color theme: light or dark (use -list-themes to see available options)")
	listCodeThemes := flag.Bool("list-code-themes", false, "List available code syntax highlighting themes and exit")
	listThemes := flag.Bool("list-themes", false, "List available PDF themes and exit")
	quiet := flag.Bool("quiet", false, "Suppress diagnostic warnings (slide overflow, code truncation)")
	showVersion := flag.Bool("version", false, "Show version information and exit")
	flag.Parse()

	// If version flag is set, print version and exit
	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// If list-themes flag is set, print available themes and exit
	if *listThemes {
		themes := converter.GetAvailableThemes()
		fmt.Println("Available PDF themes:")
		for _, theme := range themes {
			fmt.Printf("  - %s\n", theme)
		}
		os.Exit(0)
	}

	// If list-code-themes flag is set, print available themes and exit
	if *listCodeThemes {
		themes := converter.GetAvailableStyles()
		fmt.Println("Available code syntax highlighting themes:")
		for _, theme := range themes {
			fmt.Printf("  - %s\n", theme)
		}
		os.Exit(0)
	}

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: input file is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check if input file exists
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: input file does not exist: %s\n", *inputFile)
		os.Exit(1)
	}

	// Default output file
	output := *outputFile
	if output == "" {
		ext := filepath.Ext(*inputFile)
		output = (*inputFile)[:len(*inputFile)-len(ext)] + ".pdf"
	}

	// Convert slide to PDF
	conv := converter.NewConverter(
		converter.WithCodeTheme(*codeTheme),
		converter.WithTheme(*pdfTheme),
		converter.WithQuiet(*quiet),
	)
	if err := conv.Convert(*inputFile, output); err != nil {
		fmt.Fprintf(os.Stderr, "Error converting file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %s to %s\n", *inputFile, output)
}
