# Document Classification Tool

A proof-of-concept tool for processing PDF documents and analyzing security classification markings using go-agents vision capabilities. This tool validates document processing patterns for the future go-agents-document-context library.

## Overview

The classify-docs tool demonstrates document processing → agent analysis architecture by:

1. **Processing PDFs** - Extracting pages and converting to images for vision API consumption
2. **Validating Patterns** - Testing sequential processing with context accumulation for both classification and prompt generation
3. **Informing Design** - Prototyping interfaces and patterns for future document processing libraries

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

## Commands

The classify-docs tool provides two main commands integrated into `main.go`:

### generate-prompt - System Prompt Generation

Generate classification system prompts from reference policy documents using sequential processing with context accumulation.

**Usage:**

```bash
go run . generate-prompt [options]
```

**Flags:**
- `--config` (default: "config.classify-gpt4o-key.json") - Path to agent configuration file
- `--token` - API token (overrides token in config file)
- `--references` (default: "_context") - Directory containing reference PDF documents
- `--no-cache` - Disable cache usage (force regeneration)
- `--timeout` (default: "30m") - Operation timeout

**Example:**

```bash
export AZURE_API_KEY="your-api-key"
go run . generate-prompt --token $AZURE_API_KEY
```

**Sample Output:**

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
[Generated system prompt content]
---
```

The generated system prompt is saved to `.cache/system-prompt.json` with metadata including timestamp and reference document list.

### classify - Document Classification

Classify security markings in PDF documents using the generated system prompt and vision capabilities.

**Usage:**

```bash
go run . classify [options]
```

**Flags:**
- `--config` (default: "config.classify-o4-mini.json") - Path to agent configuration file
- `--token` - API token (overrides token in config file)
- `--input` - Directory containing PDF documents to classify
- `--output` (default: "classification-results.json") - Output JSON file path
- `--system-prompt` (default: ".cache/system-prompt.json") - Path to cached system prompt
- `--timeout` (default: "15m") - Operation timeout

**Example:**

```bash
go run . classify --token $AZURE_API_KEY --input _context/marked-documents
```

**Sample Output:**

```json
[
  {
    "file": "marked-documents_6.pdf",
    "classification": "SECRET//NOFORN",
    "confidence": "HIGH",
    "markings_found": ["SECRET//NOFORN"],
    "classification_rationale": "Banner marking 'SECRET//NOFORN' clearly visible..."
  },
  {
    "file": "marked-documents_19.pdf",
    "classification": "SECRET",
    "confidence": "MEDIUM",
    "markings_found": ["SECRET"],
    "classification_rationale": "A clear classification banner 'SECRET' appears..."
  }
]
```

**Accuracy Note:** Currently achieves 96.3% accuracy (26/27 documents) on test set. Known limitation: Document 19 (marked-documents_19.pdf) - the model cannot consistently detect the faded NOFORN caveat stamp, resulting in classification as 'SECRET' instead of 'SECRET//NOFORN'. This is correctly flagged with MEDIUM confidence for human review.

### Utility Tools

**test-config** - Configuration verification utility:

```bash
go run ./cmd/test-config/main.go <config-file>
```

**test-render** - PDF rendering verification utility:

```bash
cd tools/classify-docs
go build -o test-render ./cmd/test-render
./test-render --input <pdf-path> [options]
```

See source files for detailed usage information.

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

### Phase 5: Complete ✅

**Document Classification** - Sequential processing with conservative confidence scoring

- Per-page classification with context accumulation
- Comprehensive classification prompt with self-check verification
- Conservative confidence scoring (HIGH/MEDIUM/LOW)
- Suspicion-based confidence for documents with missing caveats
- Optimized for o4-mini visual reasoning model
- Achieved 96.3% accuracy (26/27 documents)

See: [04-document-classification.md](./_context/.archive/04-document-classification.md)

### Phase 6: Complete ✅

**Testing & Validation** - Validation completed through 27-document test set achieving 96.3% accuracy. Suspicion-based confidence scoring successfully flags edge cases for human review. Prototype validation complete, ready for component extraction.

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
2. **Processing Patterns** - Sequential document workflows with context accumulation and retry logic (Phase 2 - ✅ Complete)
3. **Use-Case Implementations** - System prompt generation and document classification using sequential processing (Phases 4 & 5 - ✅ Complete)

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
├── classify/      # Document classification logic
└── processing/    # Sequential processor with context accumulation
```

See [PROJECT.md](./PROJECT.md) for detailed architecture documentation.

## License

Part of the go-agents project. See root LICENSE file for details.
