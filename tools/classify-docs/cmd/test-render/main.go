package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JaimeStill/go-agents/tools/classify-docs/document"
)

func main() {
	// Define flags
	input := flag.String("input", "", "Path to PDF file (required)")
	page := flag.Int("page", 1, "Page number to render (1-indexed)")
	output := flag.String("output", "", "Output filename (default: [pdf-name].[page].[format])")
	path := flag.String("path", ".", "Output directory path")
	format := flag.String("format", "png", "Image format (png or jpeg)")
	dpi := flag.Int("dpi", 150, "Rendering DPI")
	quality := flag.Int("quality", 85, "JPEG quality (1-100, ignored for PNG)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Test PDF page rendering to images with configurable options.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Render page 1 of document.pdf to PNG\n")
		fmt.Fprintf(os.Stderr, "  %s -input document.pdf\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Render page 5 to JPEG with custom quality\n")
		fmt.Fprintf(os.Stderr, "  %s -input document.pdf -page 5 -format jpeg -quality 95\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Render to specific output file\n")
		fmt.Fprintf(os.Stderr, "  %s -input document.pdf -output test-page.png\n\n", os.Args[0])
	}

	flag.Parse()

	// Validate required flags
	if *input == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate format
	var imgFormat document.ImageFormat
	switch strings.ToLower(*format) {
	case "png":
		imgFormat = document.PNG
	case "jpeg", "jpg":
		imgFormat = document.JPEG
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid format '%s' (must be png or jpeg)\n", *format)
		os.Exit(1)
	}

	// Validate quality range
	if *quality < 1 || *quality > 100 {
		fmt.Fprintf(os.Stderr, "Error: quality must be 1-100, got %d\n", *quality)
		os.Exit(1)
	}

	// Validate DPI
	if *dpi < 72 || *dpi > 600 {
		fmt.Fprintf(os.Stderr, "Error: DPI must be 72-600, got %d\n", *dpi)
		os.Exit(1)
	}

	// Run the rendering
	if err := renderPage(*input, *page, *output, *path, imgFormat, *dpi, *quality); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func renderPage(
	inputPath string,
	pageNum int,
	outputName string,
	outputPath string,
	format document.ImageFormat,
	dpi int,
	quality int,
) error {
	// Open PDF
	fmt.Printf("Opening PDF: %s\n", inputPath)
	doc, err := document.OpenPDF(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	fmt.Printf("PDF has %d pages\n", doc.PageCount())

	// Validate page number
	if pageNum < 1 || pageNum > doc.PageCount() {
		return fmt.Errorf("page %d out of range [1-%d]", pageNum, doc.PageCount())
	}

	// Extract page
	fmt.Printf("Extracting page %d...\n", pageNum)
	page, err := doc.ExtractPage(pageNum)
	if err != nil {
		return fmt.Errorf("failed to extract page: %w", err)
	}

	// Configure image options
	opts := document.ImageOptions{
		Format:  format,
		DPI:     dpi,
		Quality: quality,
	}

	// Render page to image
	fmt.Printf("Rendering page to %s (DPI: %d", format, dpi)
	if format == document.JPEG {
		fmt.Printf(", Quality: %d", quality)
	}
	fmt.Printf(")...\n")

	imgData, err := page.ToImage(opts)
	if err != nil {
		return fmt.Errorf("failed to render page: %w", err)
	}

	// Determine output filename
	if outputName == "" {
		// Generate default: [pdf-name].[page].[format]
		pdfName := filepath.Base(inputPath)
		pdfName = strings.TrimSuffix(pdfName, filepath.Ext(pdfName))
		ext := string(format)
		outputName = fmt.Sprintf("%s.%d.%s", pdfName, pageNum, ext)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write image file
	fullOutputPath := filepath.Join(outputPath, outputName)
	if err := os.WriteFile(fullOutputPath, imgData, 0644); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}

	fmt.Printf("\nSuccess!\n")
	fmt.Printf("  Output: %s\n", fullOutputPath)
	fmt.Printf("  Size:   %d bytes (%.2f KB)\n", len(imgData), float64(len(imgData))/1024.0)
	fmt.Printf("\nOpen the image to verify rendering quality.\n")

	return nil
}
