package converter

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/tools/present"
)

const (
	imgContentX      = 20.0  // left content margin (mm)
	imgContentWidth  = 257.0 // available content width (mm)
	imgContentBottom = 190.0 // bottom boundary of slide content (mm)
)

// renderImage renders a present.Image element (.image directive, legacy format).
func (c *Converter) renderImage(img present.Image, y float64) float64 {
	imagePath := img.URL
	if !filepath.IsAbs(imagePath) {
		imagePath = filepath.Join(c.slideDir, imagePath)
	}
	return c.renderImageFile(imagePath, y)
}

// renderHTMLImage renders an <img> HTML tag from markdown-converted content.
func (c *Converter) renderHTMLImage(imgHTML string, y float64) float64 {
	srcRe := regexp.MustCompile(`(?i)src=["']([^"']+)["']`)
	match := srcRe.FindStringSubmatch(imgHTML)
	if len(match) < 2 {
		return y
	}
	imagePath := match[1]
	if !filepath.IsAbs(imagePath) {
		imagePath = filepath.Join(c.slideDir, imagePath)
	}
	return c.renderImageFile(imagePath, y)
}

// renderImageFile places an image from a file path into the PDF, centered
// horizontally and scaled to fit within the remaining slide content area.
func (c *Converter) renderImageFile(imagePath string, y float64) float64 {
	if _, err := os.Stat(imagePath); err != nil {
		if !c.quiet {
			fmt.Fprintf(os.Stderr, "Warning: slide %d %q: image not found: %s\n",
				c.currentSlideNumber, c.currentSlideTitle, imagePath)
		}
		return y
	}

	ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(imagePath), "."))
	if ext == "JPG" {
		ext = "JPEG"
	}
	switch ext {
	case "JPEG", "PNG", "GIF":
	default:
		if !c.quiet {
			fmt.Fprintf(os.Stderr, "Warning: slide %d %q: unsupported image format %q: %s\n",
				c.currentSlideNumber, c.currentSlideTitle, ext, imagePath)
		}
		return y
	}

	info := c.pdf.RegisterImageOptions(imagePath, gofpdf.ImageOptions{ImageType: ext})
	if c.pdf.Err() {
		if !c.quiet {
			fmt.Fprintf(os.Stderr, "Warning: slide %d %q: failed to load image %s: %v\n",
				c.currentSlideNumber, c.currentSlideTitle, imagePath, c.pdf.Error())
		}
		c.pdf.ClearError()
		return y
	}

	maxH := imgContentBottom - y
	if maxH <= 5 {
		return y
	}

	imgW := info.Width()
	imgH := info.Height()

	var w, h float64
	if imgW > 0 && imgH > 0 {
		scale := math.Min(imgContentWidth/imgW, maxH/imgH)
		w = imgW * scale
		h = imgH * scale
	} else {
		w = imgContentWidth
		h = 0
	}

	x := imgContentX + (imgContentWidth-w)/2
	c.pdf.ImageOptions(imagePath, x, y, w, h, false, gofpdf.ImageOptions{ImageType: ext}, 0, "")

	return y + h + 5
}
