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

### generate-prompt - System Prompt Generation

Generate classification system prompts from reference policy documents using sequential processing with context accumulation.

#### Usage

```bash
go run ./cmd/generate-prompt/main.go [options]
```

**Flags:**
- `--config` (default: "config.classify-gemma.json") - Path to agent configuration file
- `--token` - API token (overrides token in config file)
- `--references` (default: "_context") - Directory containing reference PDF documents
- `--no-cache` - Disable cache usage (force regeneration)
- `--timeout` (default: "30m") - Operation timeout

#### Examples

**Basic Usage - Generate with Default Config**

```bash
export AZURE_API_KEY="your-api-key"
go run ./cmd/generate-prompt/main.go --token $AZURE_API_KEY
```

**Custom Configuration and References**

```bash
go run ./cmd/generate-prompt/main.go \
  --config config.classify-gpt4o-key.json \
  --token $AZURE_API_KEY \
  --references _context \
  --timeout 45m
```

**Force Regeneration (Ignore Cache)**

```bash
go run ./cmd/generate-prompt/main.go --token $AZURE_API_KEY --no-cache
```

**Output:**

```
Discovering reference documents...
  Found: dodm-5200.01-enc4.pdf (9 pages)
  Found: security-classification-markings.pdf (2 pages)

Generating system prompt from 11 pages...
  Page 1/11 processed
  Page 2/11 processed
  ...
  Page 11/11 processed

System prompt generated successfully!
  Cached: .cache/system-prompt.json

---
[Generated system prompt content displayed to stdout]
---
```

The generated system prompt is saved to `.cache/system-prompt.json` with metadata including timestamp and reference document list. The prompt content is also displayed to stdout for immediate review.

### test-config - Configuration Verification

Utility for verifying configuration loading and default value merging.

#### Usage

```bash
go run ./cmd/test-config/main.go <config-file>
```

**Examples:**

```bash
# Verify Azure OpenAI config
go run ./cmd/test-config/main.go config.classify-gpt4o-key.json

# Verify Ollama config
go run ./cmd/test-config/main.go config.classify-gemma.json
```

**Output:**
```
Configuration loaded successfully from: config.classify-gpt4o-key.json

Agent Configuration:
  Name: classify-agent-gpt4o
  Provider: azure
  Model: gpt-4o
  Base URL: https://go-agents-platform.openai.azure.com/openai

Processing Configuration (with defaults):
  Parallel:
    Worker Cap: 16 (default: 16)
  Sequential:
    Expose Intermediate Contexts: false (default: false)
  Retry:
    Max Attempts: 3 (default: 3)
    Initial Backoff: 1s (default: 1s)
    Max Backoff: 30s (default: 30s)
    Backoff Multiplier: 2.0 (default: 2.0)
  Cache:
    Enabled: true (default: true)
    Path: .cache/system-prompt.json (default: .cache/system-prompt.json)
```

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

Run the complete test suite:

```bash
cd tools/classify-docs
go test ./tests/... -v
```

**Test Packages:**
- `tests/document/` - PDF processing and image conversion (7 tests)
- `tests/config/` - Configuration loading and merging (4 tests)
- `tests/cache/` - System prompt caching (4 tests)
- `tests/retry/` - Retry logic and exponential backoff (6 tests)
- `tests/processing/` - Parallel and sequential processors (12 tests)
- `tests/encoding/` - Base64 data URI image encoding (5 tests)
- `tests/prompt/` - System prompt generation validation (3 tests)

**Total: 45 tests, all passing**

Tests automatically skip image conversion tests if ImageMagick is not available.

**Run with coverage:**

```bash
go test ./tests/... -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Development Status

### Phase 1: Complete ✅

**Document Processing Primitives** - PDF extraction and image conversion infrastructure

See: [01-document-processing-primitives.md](./_context/.archive/01-document-processing-primitives.md)

### Phase 2: Complete ✅

**Processing Infrastructure** - Generic parallel and sequential processors with retry logic

- Generic processing functions with type parameters
- Parallel processor with worker pools
- Sequential processor with context accumulation
- Retry infrastructure with exponential backoff
- Unified configuration management

See: [02-processing-infrastructure.md](./_context/.archive/02-processing-infrastructure.md)

### Phase 3: Complete ✅

**Caching Infrastructure** - System prompt caching (merged into Phase 2)

- JSON-based cache persistence
- Metadata tracking (timestamp, reference documents)
- Configuration integration

See: [02-processing-infrastructure.md](./_context/.archive/02-processing-infrastructure.md)

### Phase 4: Complete ✅

**System Prompt Generation** - Sequential processing with context accumulation

- Base64 data URI encoding for vision API
- Sequential processor integration with vision agent
- Progressive system prompt refinement across document pages
- Retry infrastructure tuned for Azure rate limiting (13s initial backoff, 1.2 multiplier)
- Generic progress reporting with result visibility
- CLI tool for prompt generation

See: [03-system-prompt-generation.md](./_context/.archive/03-system-prompt-generation.md)

### Phases 5-6: Planned

- **Phase 5**: Document classification using parallel processing
- **Phase 6**: Integration testing and validation

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

1. **Document Processing Primitives** - Low-level PDF operations and image conversion (Phase 1 - ✅ Complete)
2. **Processing Patterns** - Parallel and sequential document workflows with retry logic (Phase 2 - ✅ Complete)
3. **Use-Case Implementations** - Classification and prompt generation (Phase 4 - ✅ Complete for prompt generation, Phase 5 - Planned for classification)

Supporting infrastructure includes:
- System prompt caching (Phase 3 - ✅ Complete)
- Unified configuration management (Phase 2 - ✅ Complete)
- Retry infrastructure with exponential backoff (Phase 2 - ✅ Complete)
- Base64 data URI encoding for vision API (Phase 4 - ✅ Complete)

**Package Structure:**
```
pkg/
├── config/        # Configuration types and loading
├── retry/         # Retry logic with exponential backoff
├── cache/         # System prompt caching
├── document/      # PDF processing and image conversion
├── encoding/      # Base64 data URI encoding for images
├── prompt/        # System prompt generation
└── processing/    # Parallel and sequential processors
```

See [PROJECT.md](./PROJECT.md) for detailed architecture documentation.

## License

Part of the go-agents project. See root LICENSE file for details.
