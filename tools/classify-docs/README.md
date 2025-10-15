# Document Classification Tool

A proof-of-concept tool for processing PDF documents and analyzing security classification markings using go-agents vision capabilities. This tool validates document processing patterns for the future go-agents-document-context library.

## Overview

The classify-docs tool demonstrates document processing → agent analysis architecture by:

1. **Processing PDFs** - Extracting pages and converting to images for vision API consumption
2. **Validating Patterns** - Testing both parallel (classification) and sequential (prompt generation) processing approaches
3. **Informing Design** - Prototyping interfaces and patterns for go-agents-document-context library

### Why This Tool?

Azure OpenAI Vision API does not support native PDF input. Documents must be converted to images before analysis. This tool provides the infrastructure for that preprocessing while validating reusable architecture patterns.

## Project Documentation

- **[PROJECT.md](./PROJECT.md)** - Complete project roadmap, architecture layers, and phase development plan
- **[_context/.archive/](._context/.archive/)** - Historical development summaries and implementation decisions

## Prerequisites

**Go 1.25.2+** - This tool uses a nested Go module

**ImageMagick v7+** - Required for PDF page rendering

```bash
# Arch Linux
sudo pacman -S imagemagick

# Ubuntu/Debian
sudo apt install imagemagick

# macOS
brew install imagemagick

# Verify installation
magick -version
```

**pdfcpu** - Installed automatically via Go modules

## Installation

```bash
cd tools/classify-docs
go mod download
```

## Available Tools

### test-render - PDF Rendering Verification

Manual testing utility for validating PDF page rendering with configurable options.

#### Building

```bash
cd tools/classify-docs
go build -o test-render ./cmd/test-render
```

#### Usage

```bash
./test-render --input <pdf-path> [options]
```

**Flags:**
- `--input` (required) - Path to PDF file
- `--page` (default: 1) - Page number to render (1-indexed)
- `--output` (default: auto) - Output filename
- `--path` (default: ".") - Output directory
- `--format` (default: "png") - Image format (png or jpeg)
- `--dpi` (default: 150) - Rendering DPI (72-600)
- `--quality` (default: 85) - JPEG quality (1-100, ignored for PNG)

#### Examples

**Basic Usage - Render First Page**

```bash
./test-render --input _context/security-classification-markings.pdf
```

```
Opening PDF: _context/security-classification-markings.pdf
PDF has 2 pages
Extracting page 1...
Note: Images rendered with opaque white backgrounds (no transparency)
Rendering page to png (DPI: 150)...

Success!
  Output: ./security-classification-markings.1.png
  Size:   241399 bytes (235.74 KB)

Open the image to verify rendering quality.
```

**Render Specific Page**

```bash
./test-render --input _context/security-classification-markings.pdf --page 2
```

**Test Different DPI Settings**

```bash
# Screen resolution (72 DPI)
./test-render --input sample.pdf --dpi 72 --output test-72dpi.png

# Balanced quality (150 DPI) - default
./test-render --input sample.pdf --dpi 150 --output test-150dpi.png

# Print quality (300 DPI)
./test-render --input sample.pdf --dpi 300 --output test-300dpi.png
```

Compare the three images to see quality vs file size tradeoffs.

**Compare PNG vs JPEG**

```bash
# PNG (lossless, larger files)
./test-render --input sample.pdf --format png --output test.png

# JPEG high quality
./test-render --input sample.pdf --format jpeg --quality 95 --output test-high.jpg

# JPEG medium quality
./test-render --input sample.pdf --format jpeg --quality 85 --output test-medium.jpg

# JPEG low quality
./test-render --input sample.pdf --format jpeg --quality 60 --output test-low.jpg
```

Examine file sizes and visual quality to determine optimal settings.

**Custom Output Location**

```bash
./test-render --input sample.pdf --output my-page.png --path ./output
```

Output: `./output/my-page.png`

**Process Multiple Pages**

```bash
# Render pages 1-5
for i in {1..5}; do
  ./test-render --input large-doc.pdf --page $i --path ./pages
done
```

Results in: `./pages/large-doc.1.png`, `./pages/large-doc.2.png`, etc.

## Testing

Run the test suite:

```bash
cd tools/classify-docs
go test ./tests/document/... -v
```

Tests automatically skip image conversion tests if ImageMagick is not available.

## Development Status

### Phase 1: Complete ✅

Document Processing Primitives - PDF extraction and image conversion infrastructure

### Phase 2-6: Planned

- Processing patterns (parallel and sequential)
- Caching infrastructure
- System prompt generation
- Document classification
- Integration testing

See [PROJECT.md](./PROJECT.md) for complete roadmap and architecture details.

## Troubleshooting

### ImageMagick Not Found

```
Error: exec: "magick": executable file not found in $PATH
```

**Solution:** Install ImageMagick (see Prerequisites section)

### PDF Not Authorized

```
Error: not authorized by security policy
```

**Solution:** Edit `/etc/ImageMagick-7/policy.xml` and ensure:

```xml
<policy domain="coder" rights="read|write" pattern="PDF" />
```

### Permission Denied

```
Error: failed to create output directory: permission denied
```

**Solution:** Ensure write permissions to the output directory

## Architecture

The tool is organized into three layers:

1. **Document Processing Primitives** - Low-level PDF operations and image conversion (Phase 1 - Complete)
2. **Processing Patterns** - Parallel and sequential document workflows (Phase 2 - Planned)
3. **Use-Case Implementations** - Classification and prompt generation (Phases 4-5 - Planned)

Supporting infrastructure includes image caching (Phase 3) and configuration management.

See [PROJECT.md](./PROJECT.md) for detailed architecture documentation.

## License

Part of the go-agents project. See root LICENSE file for details.
