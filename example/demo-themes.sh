#!/bin/bash
# Demonstration of theme capabilities in present2pdf

set -e

PRESENT2PDF="../present2pdf"
INPUT_FILE="presentation.slide"

echo "🎨 present2pdf Theme Demonstration"
echo "=================================="
echo ""

# Check for executable file
if [ ! -f "$PRESENT2PDF" ]; then
    echo "❌ present2pdf file not found. First run: go build -o present2pdf ./cmd/present2pdf"
    exit 1
fi

# Check for example file
if [ ! -f "$INPUT_FILE" ]; then
    echo "❌ $INPUT_FILE file not found"
    exit 1
fi

echo "📋 Available PDF themes:"
$PRESENT2PDF -list-themes
echo ""

echo "🎨 Creating examples with different themes..."
echo ""

# PDF themes to test
PDF_THEMES=("light" "dark")

# Code themes to test
CODE_THEMES=("monokai" "github" "dracula" "nord" "vim")

counter=1

# Generate PDFs with different theme combinations
for pdf_theme in "${PDF_THEMES[@]}"; do
    for code_theme in "${CODE_THEMES[@]}"; do
        output_file="demo-${pdf_theme}-${code_theme}.pdf"
        echo "${counter}️⃣  ${pdf_theme^} theme + ${code_theme^} code..."
        $PRESENT2PDF -input "$INPUT_FILE" -output "$output_file" -theme "$pdf_theme" -code-theme "$code_theme"
        echo "   ✅ Created: $output_file"
        echo ""
        ((counter++))
    done
done

echo "✨ Done! Created $((counter - 1)) PDF files with different themes."
echo ""
echo "📂 All files are in the current directory"
ls -lh demo-*.pdf
echo ""
echo "💡 Tip: Open the files to compare different themes!"
