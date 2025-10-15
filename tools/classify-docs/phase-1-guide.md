# Phase 1: Document Processing Primitives

## Phase Overview

### Objectives

Build the foundational document processing layer that provides:
- Clean, extensible interfaces for document and page operations
- PDF implementation using pure Go (pdfcpu)
- Page extraction and image conversion capabilities
- Proper resource management and cleanup

### Deliverables

1. `document/document.go` - Core interfaces and types
2. `document/pdf.go` - PDF implementation
3. Unit tests validating extraction and conversion

### Why This Matters

These primitives will be reused across:
- System prompt generation (sequential processing)
- Document classification (parallel processing)
- Caching infrastructure (image persistence)

The interfaces defined here will inform the design of go-agents-document-context library.

## Prerequisites

### Required Knowledge

- Go interfaces and type composition
- Resource management (defer, cleanup patterns)
- Error handling and error wrapping
- Basic image processing concepts (formats, DPI, encoding)

### Module Setup

This tool uses a nested Go module separate from the root go-agents library. This keeps the pdfcpu dependency isolated to the tool and maintains a lightweight go-agents library.

Your `tools/classify-docs/go.mod` should look like:

```go
module github.com/JaimeStill/go-agents/tools/classify-docs

go 1.25.2

require (
    github.com/JaimeStill/go-agents v0.1.2
    github.com/pdfcpu/pdfcpu v0.11.0
)
```

**Why pdfcpu?**
- Pure Go implementation (no CGo, no external dependencies)
- Well-maintained and actively developed
- Comprehensive PDF manipulation capabilities
- Clean API for page extraction and rendering

## Design Decisions

### Interface-Based Design

**Rationale**: Interfaces enable:
- Multiple format support (PDF, DOCX, images) in the future
- Easy testing with mocks
- Clear contracts between layers
- Library extraction without breaking changes

### Page-Level Abstraction

**Rationale**: Page-level operations enable:
- Parallel processing (process pages independently)
- Sequential processing (process pages in order)
- Memory efficiency (process one page at a time)
- Granular error handling (fail on specific page)

### Image Conversion at Page Level

**Rationale**: Converting pages to images at the Page level:
- Keeps conversion logic close to the data
- Allows per-page format/quality configuration
- Supports caching strategies
- Matches Azure OpenAI's vision API requirements

### Resource Cleanup via Close()

**Rationale**: Explicit cleanup:
- Prevents resource leaks in long-running processes
- Makes resource lifecycle explicit
- Follows Go io.Closer pattern
- Enables defer-based cleanup

## Implementation Steps

### Step 1: Create Package Structure

```bash
cd tools/classify-docs
mkdir -p document
```

### Step 2: Define Core Interfaces (`document/document.go`)

Create `document/document.go` with the following structure:

#### 2.1: Package Declaration and Imports

```go
package document

import (
    "fmt"
    "io"
)
```

#### 2.2: Image Format Type

Define an enum-like type for supported image formats:

```go
// ImageFormat represents supported image output formats
type ImageFormat string

const (
    // PNG format provides lossless compression with higher quality
    PNG ImageFormat = "png"

    // JPEG format provides lossy compression with smaller file sizes
    JPEG ImageFormat = "jpeg"
)
```

**Design Note**: Start with PNG and JPEG. Additional formats (TIFF, WebP) can be added later if needed.

#### 2.3: Image Options Structure

Define configuration for image conversion:

```go
// ImageOptions configures image conversion parameters
type ImageOptions struct {
    // Format specifies the output image format (PNG or JPEG)
    Format ImageFormat

    // Quality specifies JPEG compression quality (1-100, ignored for PNG)
    // Higher values produce better quality but larger files
    Quality int

    // DPI specifies the rendering resolution (dots per inch)
    // Higher DPI produces clearer images but increases file size
    // Common values: 72 (screen), 150 (balanced), 300 (print quality)
    DPI int
}

// DefaultImageOptions provides reasonable defaults for most use cases
func DefaultImageOptions() ImageOptions {
    return ImageOptions{
        Format:  PNG,    // Lossless for document fidelity
        Quality: 0,      // N/A for PNG
        DPI:     150,    // Balance between quality and size
    }
}
```

**Design Note**: DPI of 150 balances quality and file size for vision API consumption. Lower DPI (72-96) may miss fine details; higher DPI (300+) unnecessarily increases token usage.

#### 2.4: Document Interface

Define the primary document interface:

```go
// Document represents a multi-page document that can be processed page-by-page
type Document interface {
    // PageCount returns the total number of pages in the document
    PageCount() int

    // ExtractPage extracts a specific page by number (1-indexed)
    // Returns an error if the page number is out of range
    ExtractPage(pageNum int) (Page, error)

    // ExtractAllPages extracts all pages in the document
    // Returns a slice of pages in document order
    ExtractAllPages() ([]Page, error)

    // Close releases resources associated with the document
    // Should be called when done processing the document
    Close() error
}
```

**Design Note**: 1-indexed page numbers match user expectations and PDF conventions. Internally, pdfcpu uses 1-indexed pages.

#### 2.5: Page Interface

Define the page-level interface:

```go
// Page represents a single page from a document
type Page interface {
    // Number returns the page number (1-indexed)
    Number() int

    // ToImage converts the page to an image with the specified options
    // Returns the image data as a byte slice
    ToImage(opts ImageOptions) ([]byte, error)

    // Close releases resources associated with the page
    // Should be called when done processing the page
    Close() error
}
```

**Design Note**: `ToImage()` returns `[]byte` rather than `image.Image` to match how we'll send data to the vision API (base64 encoding of raw bytes).

### Step 3: Implement PDF Support (`document/pdf.go`)

Create `document/pdf.go` with PDF-specific implementation:

#### 3.1: Package and Imports

```go
package document

import (
    "bytes"
    "fmt"
    "image/jpeg"
    "image/png"

    "github.com/pdfcpu/pdfcpu/pkg/api"
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)
```

#### 3.2: PDFDocument Structure

```go
// PDFDocument implements Document interface for PDF files
type PDFDocument struct {
    path     string
    ctx      *model.Context
    pageCount int
}
```

**Design Note**: Store `*model.Context` to avoid re-reading the PDF for each operation. Cache `pageCount` for efficiency.

#### 3.3: OpenPDF Function

```go
// OpenPDF opens a PDF document from the specified file path
// Returns an error if the file cannot be read or is not a valid PDF
func OpenPDF(path string) (*PDFDocument, error) {
    // Read and parse PDF
    ctx, err := api.ReadContextFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open PDF: %w", err)
    }

    // Get page count
    pageCount := ctx.PageCount
    if pageCount == 0 {
        return nil, fmt.Errorf("PDF has no pages")
    }

    return &PDFDocument{
        path:      path,
        ctx:       ctx,
        pageCount: pageCount,
    }, nil
}
```

**Error Handling**: Wrap errors with context using `fmt.Errorf` and `%w` for error chains.

#### 3.4: Implement Document Interface Methods

```go
// PageCount returns the total number of pages
func (d *PDFDocument) PageCount() int {
    return d.pageCount
}

// ExtractPage extracts a specific page (1-indexed)
func (d *PDFDocument) ExtractPage(pageNum int) (Page, error) {
    if pageNum < 1 || pageNum > d.pageCount {
        return nil, fmt.Errorf("page %d out of range [1-%d]", pageNum, d.pageCount)
    }

    return &PDFPage{
        doc:    d,
        number: pageNum,
    }, nil
}

// ExtractAllPages extracts all pages in order
func (d *PDFDocument) ExtractAllPages() ([]Page, error) {
    pages := make([]Page, 0, d.pageCount)

    for i := 1; i <= d.pageCount; i++ {
        page, err := d.ExtractPage(i)
        if err != nil {
            return nil, fmt.Errorf("failed to extract page %d: %w", i, err)
        }
        pages = append(pages, page)
    }

    return pages, nil
}

// Close releases resources
func (d *PDFDocument) Close() error {
    // pdfcpu Context doesn't require explicit cleanup
    // This method is here for interface compliance and future-proofing
    d.ctx = nil
    return nil
}
```

**Design Note**: `ExtractPage` is cheap—it just creates a `PDFPage` wrapper. Actual rendering happens in `ToImage()`.

#### 3.5: PDFPage Structure

```go
// PDFPage implements Page interface for PDF pages
type PDFPage struct {
    doc    *PDFDocument
    number int
}

// Number returns the page number (1-indexed)
func (p *PDFPage) Number() int {
    return p.number
}

// Close releases page resources
func (p *PDFPage) Close() error {
    // No cleanup needed for page references
    return nil
}
```

#### 3.6: Implement ToImage (Core Conversion Logic)

This is the most complex method—it renders a PDF page to an image:

```go
// ToImage converts the PDF page to an image
func (p *PDFPage) ToImage(opts ImageOptions) ([]byte, error) {
    // Use default options if DPI is not specified
    if opts.DPI == 0 {
        opts = DefaultImageOptions()
    }

    // Render PDF page to image using pdfcpu
    img, err := api.PageImage(p.doc.ctx, p.number, opts.DPI)
    if err != nil {
        return nil, fmt.Errorf("failed to render page %d: %w", p.number, err)
    }

    // Encode image to specified format
    var buf bytes.Buffer

    switch opts.Format {
    case PNG:
        err = png.Encode(&buf, img)
        if err != nil {
            return nil, fmt.Errorf("failed to encode PNG: %w", err)
        }

    case JPEG:
        quality := opts.Quality
        if quality == 0 {
            quality = 85  // Default JPEG quality
        }
        if quality < 1 || quality > 100 {
            return nil, fmt.Errorf("JPEG quality must be 1-100, got %d", quality)
        }

        err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
        if err != nil {
            return nil, fmt.Errorf("failed to encode JPEG: %w", err)
        }

    default:
        return nil, fmt.Errorf("unsupported image format: %s", opts.Format)
    }

    return buf.Bytes(), nil
}
```

**Implementation Notes**:
- `api.PageImage()` renders the PDF page at specified DPI
- PNG encoding is lossless, no quality parameter needed
- JPEG quality defaults to 85 (good balance)
- Errors are wrapped with context for debugging

**Performance Note**: This is the expensive operation. Rendering a page at 150 DPI typically takes 200-500ms depending on page complexity.

### Step 4: Create Test File (`document/document_test.go`)

#### 4.1: Test Package Setup

```go
package document_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/JaimeStill/go-agents/tools/classify-docs/document"
)
```

**Design Note**: Black-box testing using `package document_test` ensures we test the public API.

#### 4.2: Test Helper - Sample PDF Path

```go
// testPDFPath returns path to a test PDF
// For now, use one of your sample documents
func testPDFPath(t *testing.T) string {
    t.Helper()

    // Use the classification guide as test document
    path := filepath.Join("..", "_context", "security-classification-markings.pdf")

    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Fatalf("Test PDF not found: %s", path)
    }

    return path
}
```

#### 4.3: Test OpenPDF

```go
func TestOpenPDF(t *testing.T) {
    path := testPDFPath(t)

    doc, err := document.OpenPDF(path)
    if err != nil {
        t.Fatalf("OpenPDF failed: %v", err)
    }
    defer doc.Close()

    if doc.PageCount() == 0 {
        t.Error("Expected non-zero page count")
    }

    t.Logf("Successfully opened PDF with %d pages", doc.PageCount())
}

func TestOpenPDF_InvalidPath(t *testing.T) {
    _, err := document.OpenPDF("/nonexistent/file.pdf")
    if err == nil {
        t.Error("Expected error for nonexistent file")
    }
}
```

#### 4.4: Test Page Extraction

```go
func TestPDFDocument_ExtractPage(t *testing.T) {
    path := testPDFPath(t)
    doc, err := document.OpenPDF(path)
    if err != nil {
        t.Fatalf("OpenPDF failed: %v", err)
    }
    defer doc.Close()

    // Test valid page
    page, err := doc.ExtractPage(1)
    if err != nil {
        t.Fatalf("ExtractPage failed: %v", err)
    }

    if page.Number() != 1 {
        t.Errorf("Expected page number 1, got %d", page.Number())
    }

    // Test out of range
    _, err = doc.ExtractPage(0)
    if err == nil {
        t.Error("Expected error for page 0")
    }

    _, err = doc.ExtractPage(doc.PageCount() + 1)
    if err == nil {
        t.Error("Expected error for page beyond document")
    }
}

func TestPDFDocument_ExtractAllPages(t *testing.T) {
    path := testPDFPath(t)
    doc, err := document.OpenPDF(path)
    if err != nil {
        t.Fatalf("OpenPDF failed: %v", err)
    }
    defer doc.Close()

    pages, err := doc.ExtractAllPages()
    if err != nil {
        t.Fatalf("ExtractAllPages failed: %v", err)
    }

    if len(pages) != doc.PageCount() {
        t.Errorf("Expected %d pages, got %d", doc.PageCount(), len(pages))
    }

    // Verify page numbers are sequential
    for i, page := range pages {
        expectedNum := i + 1
        if page.Number() != expectedNum {
            t.Errorf("Page %d has wrong number: %d", i, page.Number())
        }
    }
}
```

#### 4.5: Test Image Conversion

```go
func TestPDFPage_ToImage_PNG(t *testing.T) {
    path := testPDFPath(t)
    doc, err := document.OpenPDF(path)
    if err != nil {
        t.Fatalf("OpenPDF failed: %v", err)
    }
    defer doc.Close()

    page, err := doc.ExtractPage(1)
    if err != nil {
        t.Fatalf("ExtractPage failed: %v", err)
    }

    opts := document.ImageOptions{
        Format: document.PNG,
        DPI:    150,
    }

    imgData, err := page.ToImage(opts)
    if err != nil {
        t.Fatalf("ToImage failed: %v", err)
    }

    if len(imgData) == 0 {
        t.Error("Expected non-empty image data")
    }

    // PNG files start with specific magic bytes
    if len(imgData) < 8 || imgData[0] != 0x89 || imgData[1] != 'P' {
        t.Error("Image data does not appear to be PNG format")
    }

    t.Logf("Generated PNG: %d bytes", len(imgData))
}

func TestPDFPage_ToImage_JPEG(t *testing.T) {
    path := testPDFPath(t)
    doc, err := document.OpenPDF(path)
    if err != nil {
        t.Fatalf("OpenPDF failed: %v", err)
    }
    defer doc.Close()

    page, err := doc.ExtractPage(1)
    if err != nil {
        t.Fatalf("ExtractPage failed: %v", err)
    }

    opts := document.ImageOptions{
        Format:  document.JPEG,
        Quality: 85,
        DPI:     150,
    }

    imgData, err := page.ToImage(opts)
    if err != nil {
        t.Fatalf("ToImage failed: %v", err)
    }

    if len(imgData) == 0 {
        t.Error("Expected non-empty image data")
    }

    // JPEG files start with 0xFF 0xD8
    if len(imgData) < 2 || imgData[0] != 0xFF || imgData[1] != 0xD8 {
        t.Error("Image data does not appear to be JPEG format")
    }

    t.Logf("Generated JPEG: %d bytes", len(imgData))
}

func TestPDFPage_ToImage_DefaultOptions(t *testing.T) {
    path := testPDFPath(t)
    doc, err := document.OpenPDF(path)
    if err != nil {
        t.Fatalf("OpenPDF failed: %v", err)
    }
    defer doc.Close()

    page, err := doc.ExtractPage(1)
    if err != nil {
        t.Fatalf("ExtractPage failed: %v", err)
    }

    // Pass zero-value options to test defaults
    imgData, err := page.ToImage(document.ImageOptions{})
    if err != nil {
        t.Fatalf("ToImage with defaults failed: %v", err)
    }

    if len(imgData) == 0 {
        t.Error("Expected non-empty image data")
    }
}
```

#### 4.6: Run Tests

```bash
cd tools/classify-docs
go test ./document/... -v
```

Expected output:
```
=== RUN   TestOpenPDF
--- PASS: TestOpenPDF (0.05s)
=== RUN   TestOpenPDF_InvalidPath
--- PASS: TestOpenPDF_InvalidPath (0.00s)
=== RUN   TestPDFDocument_ExtractPage
--- PASS: TestPDFDocument_ExtractPage (0.01s)
=== RUN   TestPDFDocument_ExtractAllPages
--- PASS: TestPDFDocument_ExtractAllPages (0.02s)
=== RUN   TestPDFPage_ToImage_PNG
--- PASS: TestPDFPage_ToImage_PNG (0.45s)
=== RUN   TestPDFPage_ToImage_JPEG
--- PASS: TestPDFPage_ToImage_JPEG (0.42s)
=== RUN   TestPDFPage_ToImage_DefaultOptions
--- PASS: TestPDFPage_ToImage_DefaultOptions (0.43s)
PASS
```

## Testing Strategy

### Unit Tests

Focus on:
- ✅ Interface contract compliance
- ✅ Error handling for edge cases
- ✅ Resource cleanup behavior
- ✅ Image format validation
- ✅ Page range validation

### Manual Testing

Create a simple test program to validate image quality:

```go
// tools/classify-docs/test-image-quality/main.go
package main

import (
    "fmt"
    "os"

    "github.com/JaimeStill/go-agents/tools/classify-docs/document"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go <pdf-path>")
        os.Exit(1)
    }

    doc, err := document.OpenPDF(os.Args[1])
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    defer doc.Close()

    fmt.Printf("Processing %d pages...\n", doc.PageCount())

    page, _ := doc.ExtractPage(1)

    opts := document.DefaultImageOptions()
    imgData, err := page.ToImage(opts)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    outPath := "test-page-1.png"
    err = os.WriteFile(outPath, imgData, 0644)
    if err != nil {
        fmt.Printf("Error writing file: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Wrote %s (%d bytes)\n", outPath, len(imgData))
    fmt.Println("Manually review the image to verify quality")
}
```

Run and review the generated image for quality.

## Verification Checklist

Before proceeding to Phase 2, verify:

- [ ] `document/document.go` compiles without errors
- [ ] `document/pdf.go` compiles without errors
- [ ] All tests pass: `go test ./document/... -v`
- [ ] Can open PDF files from disk
- [ ] Can extract individual pages by number
- [ ] Can extract all pages in order
- [ ] Can convert pages to PNG images
- [ ] Can convert pages to JPEG images
- [ ] Default image options work correctly
- [ ] Out-of-range page numbers return errors
- [ ] Invalid PDF paths return errors
- [ ] Generated images are visually correct (manual check)
- [ ] Resources are cleaned up properly (no panics, no leaks)

## Common Pitfalls

### 1. Page Numbering Confusion

**Issue**: Off-by-one errors mixing 0-indexed and 1-indexed page numbers

**Solution**: Use 1-indexed consistently in public API. Document this clearly. Add validation in `ExtractPage()`.

### 2. Large Memory Usage

**Issue**: Keeping all rendered images in memory

**Solution**: Phase 1 doesn't address this yet. In future phases, use caching and process pages incrementally.

### 3. pdfcpu API Changes

**Issue**: pdfcpu occasionally changes APIs between versions

**Solution**: Lock pdfcpu version in `go.mod` until architecture is stable. Document which version is tested.

### 4. Image Quality vs. Size

**Issue**: PNG files can be very large at high DPI

**Solution**: 150 DPI is a good default. Let users configure via `ImageOptions`. Monitor Azure's 20MB limit in Phase 5.

## Performance Considerations

### Current Implementation

- Opening a PDF: ~50-100ms (one-time cost)
- Extracting a page reference: ~1μs (cheap, no rendering)
- Rendering page to image: ~200-500ms per page (expensive)

### Future Optimizations (Not in Phase 1)

- Parallel page rendering
- Incremental processing (don't load all pages at once)
- Image caching (Phase 3)
- Resolution adjustment based on content complexity

## Next Steps

After completing Phase 1:

1. Review generated images manually for quality
2. Confirm interfaces feel natural to use
3. Document any issues or awkward patterns
4. Request `phase-2-guide.md` for processing patterns implementation

### Phase 2 Preview

Phase 2 will build on these primitives to create:
- **Parallel Processor**: Process pages concurrently using worker pools
- **Sequential Processor**: Process pages in order with context accumulation

Both processors will use the `Document` and `Page` interfaces defined in Phase 1.

## Troubleshooting

### Test PDF Not Found

If tests fail with "Test PDF not found", ensure:
- You're running tests from `tools/classify-docs` directory
- The `_context/` directory exists
- Sample PDFs are present in `_context/`

### pdfcpu Errors

If you see pdfcpu-related errors:
- Check PDF is valid and not corrupted
- Try opening PDF in a PDF viewer first
- Check pdfcpu version: `go list -m github.com/pdfcpu/pdfcpu`

### Image Encoding Errors

If image encoding fails:
- Verify format is PNG or JPEG
- Check JPEG quality is in range 1-100
- Ensure sufficient memory for rendering

## Design Reflections

After implementing Phase 1, consider:

1. **Are the interfaces sufficient?** - Do they support both parallel and sequential processing needs?

2. **Is resource management clean?** - Is the Close() pattern working well?

3. **Are errors actionable?** - Do error messages help diagnose issues?

4. **Is the API intuitive?** - Would a new user understand how to use these interfaces?

5. **What's missing?** - Are there operations we'll need in later phases that should be added now?

Document your answers—they'll inform go-agents-document-context design.
