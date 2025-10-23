# Phase 1: Document Processing Primitives - Development Summary

## Starting Point

This phase established the foundational document processing layer for the classify-docs tool as part of validating architecture patterns for the future go-agents-document-context library. The goal was to create clean interfaces for document and page operations with PDF implementation using pure Go.

### Initial Requirements

- Process PDF documents page-by-page for Azure OpenAI vision API (which doesn't support native PDF input)
- Extract pages and convert to images (PNG/JPEG)
- Support both parallel processing (classification) and sequential processing (prompt generation)
- Clean resource management and error handling
- Extensible interfaces for future format support

## Implementation Decisions

### Technology Choices

**pdfcpu for PDF Structure Operations**
- Pure Go implementation (no CGo dependencies)
- Well-maintained library with clean API
- Excellent for PDF parsing, validation, and structure operations
- Provides access to page count and document metadata

**ImageMagick for Page Rendering**
- Industry-standard image manipulation tool
- High-quality PDF page rendering via Ghostscript
- Simple command-line interface via `magick` command
- Multi-format support (PNG, JPEG)
- Well-suited for containerized deployments

**Hybrid Approach Rationale**: Using pdfcpu for structure (fast, pure Go) and ImageMagick for rendering (battle-tested, reliable) provides optimal performance and quality.

### Interface Design

**Document Interface**
```go
type Document interface {
    PageCount() int
    ExtractPage(pageNum int) (Page, error)
    ExtractAllPages() ([]Page, error)
    Close() error
}
```

**Page Interface (Final)**
```go
type Page interface {
    Number() int
    ToImage(opts ImageOptions) ([]byte, error)
}
```

**Design Changes**: Originally included `Close() error` on Page interface but removed it after implementation showed it was unnecessary. Page references require no cleanup—only the parent Document needs explicit cleanup.

### Image Format Handling

**Format Constants**
```go
const (
    PNG  ImageFormat = "png"
    JPEG ImageFormat = "jpg"  // Constant named JPEG, value "jpg" for file extension
)
```

**Rationale**: Compromise naming where constant uses proper format name (JPEG) but value uses common file extension (jpg).

**Default Options**
- Format: PNG (lossless for document fidelity)
- DPI: 150 (balance between quality and file size)
- JPEG Quality: 85 (when JPEG is used)

### Page Numbering Convention

**Public API**: 1-indexed page numbers (matches user expectations and PDF conventions)

**Internal Conversion**: ImageMagick uses 0-based indexing, conversion handled in `ToImage()` method:
```go
pageIndex := p.number - 1
inputSpec := fmt.Sprintf("%s[%d]", p.doc.path, pageIndex)
```

### ImageMagick Command Construction

**Critical Implementation Details**:
```go
args := []string{
    "-density", fmt.Sprintf("%d", opts.DPI),
    inputSpec,                 // Input file MUST come after density
    "-background", "white",    // Set white background for transparency
    "-flatten",                // Flatten image layers onto background
}
```

**Command Order Requirements**:
1. `-density` flag must precede input file specification
2. Input file specified as `path[pageIndex]` for page selection
3. Transformation operations (`-background`, `-flatten`) come after input
4. Output path specified last

**Transparency Handling**: PDFs often contain transparency information that ImageMagick preserves by default. The `-flatten` operation composites the image onto a white background, ensuring opaque output suitable for vision API processing.

**Format-Specific Options**: JPEG quality parameter appended only for JPEG format:
```go
if opts.Format == JPEG {
    args = append(args, "-quality", fmt.Sprintf("%d", opts.Quality))
}
```

### Resource Management

**Temporary File Strategy**:
- Create temporary file for ImageMagick output using `os.CreateTemp()`
- Clean up immediately after reading with `defer os.Remove(tmpPath)`
- Temporary files scoped to single `ToImage()` call

**Document Lifecycle**:
- `OpenPDF()` reads and parses PDF structure once
- Document maintains `*model.Context` for page operations
- `Close()` releases pdfcpu context (sets to nil)
- Page extraction is cheap (creates lightweight wrapper, no rendering)

## Technical Challenges and Solutions

### Challenge 1: Transparent Backgrounds

**Problem**: Initial implementation produced images with transparent backgrounds where white should be, making text unreadable.

**Cause**: ImageMagick preserves PDF transparency by default.

**Failed Approach**: Used `-alpha remove` and `-alpha off` flags, which caused error:
```
magick: no images found for operation '-alpha' at CLI arg 5
```

**Root Cause of Failure**: Alpha flags were applied before image was loaded.

**Final Solution**: Use `-flatten` operation after loading image:
```go
args := []string{
    "-density", fmt.Sprintf("%d", opts.DPI),
    inputSpec,              // Input first
    "-background", "white", // Then background
    "-flatten",             // Then flatten
}
```

**Outcome**: Single `-flatten` operation correctly removes transparency by compositing onto white background.

### Challenge 2: ImageMagick Command Deprecation

**Issue**: ImageMagick v7 shows warning about deprecated `convert` command.

**Solution**: Updated all implementations to use `magick` command instead of deprecated `convert`.

### Challenge 3: Format Constant Naming

**Discussion**: Tension between proper format name (JPEG) and common file extension (jpg).

**Resolution**: Named constant `JPEG` with value `"jpg"` - provides semantic clarity in code while using standard file extension in filesystem operations.

## Testing Strategy

### Unit Tests

**Location**: `tests/document/pdf_test.go`

**Approach**: Black-box testing using `package document_test` to test public API only.

**Test Coverage**:
- PDF opening and validation
- Invalid path handling
- Page extraction (individual and all pages)
- Page range validation
- PNG image conversion
- JPEG image conversion
- Default options handling
- Format validation (magic byte checking)

**ImageMagick Dependency Handling**:
```go
func requireImageMagick(t *testing.T) {
    t.Helper()
    if !hasImageMagick() {
        t.Skip("ImageMagick not installed, skipping image conversion test")
    }
}
```

Tests gracefully skip when ImageMagick unavailable rather than failing.

**Test Results**: All 7 tests pass in ~1.7 seconds.

### Manual Testing Tool

**Implementation**: Created `cmd/test-render/main.go` CLI tool for manual verification of image quality.

**Flags**:
- `--input` (required): PDF path
- `--page` (default: 1): Page number to render
- `--output` (default: `[pdf-name].[page].[format]`): Output filename
- `--path` (default: "."): Output directory
- `--format` (default: "png"): Image format
- `--dpi` (default: 150): Rendering DPI (72-600)
- `--quality` (default: 85): JPEG quality (1-100)

**Usage Example**:
```bash
cd tools/classify-docs
go build -o test-render ./cmd/test-render
./test-render -input _context/security-classification-markings.pdf
```

**Output**:
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

**Testing Scenarios Supported**:
- DPI comparison (72 vs 150 vs 300)
- Format comparison (PNG vs JPEG at various quality settings)
- Multi-page document verification
- Custom output paths and names

## Final Architecture

### Package Structure

```
tools/classify-docs/
├── document/
│   ├── document.go          # Core interfaces and types
│   └── pdf.go               # PDF implementation
├── tests/
│   └── document/
│       └── pdf_test.go      # Black-box tests
├── cmd/
│   └── test-render/
│       └── main.go          # Manual testing tool
├── _context/
│   └── .archive/
│       └── 01-document-processing-primitives.md  # This document
├── go.mod                   # Nested module configuration
├── PROJECT.md               # Project roadmap
└── README.md                # User documentation
```

### Public API Surface

**Types**:
- `Document` interface
- `Page` interface
- `ImageFormat` type with `PNG` and `JPEG` constants
- `ImageOptions` struct

**Functions**:
- `OpenPDF(path string) (*PDFDocument, error)`
- `DefaultImageOptions() ImageOptions`

**Concrete Implementations**:
- `PDFDocument` (implements `Document`)
- `PDFPage` (implements `Page`)

### Dependencies

**Go Dependencies** (go.mod):
```
require github.com/pdfcpu/pdfcpu v0.11.0
```

**External Dependencies**:
- ImageMagick v7+ (`magick` command)
- Ghostscript (ImageMagick dependency for PDF rendering)

## Design Validation Results

### Interface Sufficiency

**Assessment**: Interfaces are sufficient for both parallel and sequential processing patterns.

**Rationale**:
- `ExtractPage()` enables parallel processing (pages processed independently)
- `ExtractAllPages()` enables sequential processing (pages processed in order)
- Page interface provides conversion without constraining usage patterns

### Resource Management

**Assessment**: Clean resource management with one simplification.

**Findings**:
- Document-level cleanup is necessary and well-defined
- Page-level cleanup proved unnecessary (removed `Page.Close()`)
- Page references are lightweight and require no explicit cleanup
- Temporary files properly managed with defer statements

### Error Handling

**Assessment**: Errors are actionable and informative.

**Error Message Patterns**:
- Wrapped errors with context (`fmt.Errorf` with `%w`)
- Specific page numbers in error messages
- ImageMagick output included in render failures
- Range validation errors specify valid range

**Example**:
```
page 999 out of range [1-2]
imagemagick failed for page 1: ...\nOutput: ...
```

### API Intuitiveness

**Assessment**: API is natural and follows Go conventions.

**Observations**:
- 1-indexed page numbers match user expectations
- `ToImage()` method clearly indicates conversion operation
- Options struct pattern is idiomatic Go
- Default options function provides sensible starting point

### Missing Functionality

**Assessment**: No critical functionality missing for MVP.

**Future Considerations** (for go-agents-document-context):
- Batch processing optimizations
- Image size validation (Azure 20MB limit)
- Memory-efficient streaming for large documents
- Additional format support (DOCX, XLSX, images)
- Caching infrastructure (Phase 3)

## Performance Characteristics

**Measured Performance** (security-classification-markings.pdf, 2-page document):
- Opening PDF: ~10-50ms (one-time cost)
- Extracting page reference: <1ms (cheap wrapper creation)
- Rendering page to PNG at 150 DPI: ~400-600ms per page
- Rendering page to JPEG at 150 DPI: ~400-500ms per page

**Performance Notes**:
- ImageMagick subprocess overhead: ~10-20ms per invocation
- PNG files larger than JPEG (lossless vs lossy compression)
- DPI significantly impacts both render time and file size
- No caching in Phase 1 (addressed in Phase 3)

## Lessons Learned

### What Worked Well

1. **Hybrid Approach**: Combining pdfcpu (Go) and ImageMagick (external) provided best of both worlds
2. **Interface-First Design**: Defining interfaces before implementation clarified requirements
3. **Graceful Test Skipping**: Tests that skip when ImageMagick unavailable prevent false failures
4. **Manual Testing Tool**: Dedicated CLI tool essential for validating image quality visually
5. **Incremental Problem Solving**: Transparency issue resolved through iteration and experimentation

### What Required Iteration

1. **Transparency Handling**: Multiple attempts required to find correct ImageMagick flags
2. **Format Naming**: Needed discussion to balance semantic clarity with pragmatic file extensions
3. **Close() Method**: Initially included on Page interface but removed after proving unnecessary

### Recommendations for go-agents-document-context

1. **Interface Design**: Current Document/Page interfaces are solid foundation for library
2. **Format Extensibility**: ImageFormat type easily extended for additional formats
3. **Cleanup Pattern**: Document-level cleanup sufficient; page-level cleanup not needed
4. **Error Context**: Always include page numbers and operation details in errors
5. **Testing Strategy**: Separate optional dependency tests (ImageMagick) from core logic tests
6. **Manual Validation**: Provide testing tools for visual quality verification

## Current State

### Completed

- ✅ Core interfaces defined (`Document`, `Page`)
- ✅ PDF implementation using pdfcpu
- ✅ Image conversion using ImageMagick (PNG and JPEG)
- ✅ Configurable image options (format, DPI, quality)
- ✅ Comprehensive unit tests (7 tests, all passing)
- ✅ Manual testing tool (`cmd/test-render`)
- ✅ Resource cleanup patterns established
- ✅ Error handling with context
- ✅ Page.Close() removed (simplified interface)

### Known Limitations

- No caching (pages re-rendered on each ToImage() call)
- No parallel rendering (sequential only)
- No image size validation (Azure 20MB limit not enforced)
- No batch optimizations
- Single format support (PDF only)

### Files Delivered

1. **`document/document.go`** - Core interfaces and types (36 lines)
2. **`document/pdf.go`** - PDF implementation (147 lines)
3. **`tests/document/pdf_test.go`** - Unit tests (194 lines)
4. **`cmd/test-render/main.go`** - Manual testing tool (184 lines)
5. **`_context/.archive/01-document-processing-primitives.md`** - This document

### Next Phase Prerequisites

Phase 2 (Processing Infrastructure) depends on:
- Stable Document/Page interfaces (✅ complete)
- Working image conversion (✅ complete)
- Validated resource management patterns (✅ complete)

Phase 2 will implement:
- Parallel processor with worker pools
- Sequential processor with context accumulation
- Result aggregation patterns
- Error handling strategies

## Reference Information

### ImageMagick Command Reference

**Basic Page Render**:
```bash
magick -density 150 input.pdf[0] -background white -flatten output.png
```

**JPEG with Quality**:
```bash
magick -density 150 input.pdf[0] -background white -flatten -quality 85 output.jpg
```

**Flag Order Requirements**:
1. `-density` before input file
2. Input file specification
3. Transformation operations (`-background`, `-flatten`)
4. Format-specific options (`-quality`)
5. Output path

### Common Issues and Solutions

**Issue**: Transparent backgrounds in output
**Solution**: Use `-flatten` operation after `-background white`

**Issue**: `magick: no images found for operation '-alpha'`
**Solution**: Don't use `-alpha` flags; use `-flatten` instead

**Issue**: ImageMagick not found in PATH
**Solution**: Install ImageMagick: `sudo pacman -S imagemagick`

**Issue**: PDF not authorized by security policy
**Solution**: Edit `/etc/ImageMagick-7/policy.xml` to allow PDF coder:
```xml
<policy domain="coder" rights="read|write" pattern="PDF" />
```

### Testing Commands

**Run Unit Tests**:
```bash
cd tools/classify-docs
go test ./tests/document/... -v
```

**Build Manual Testing Tool**:
```bash
cd tools/classify-docs
go build -o test-render ./cmd/test-render
```

**Render Test Page**:
```bash
./test-render -input _context/security-classification-markings.pdf -page 1
```

**Test Different Settings**:
```bash
# High DPI PNG
./test-render -input sample.pdf -dpi 300 -format png

# High quality JPEG
./test-render -input sample.pdf -format jpeg -quality 95

# Specific output
./test-render -input sample.pdf -output test.png -path ./output
```

## Conclusion

Phase 1 successfully established clean, extensible document processing primitives that will serve as the foundation for both parallel and sequential processing patterns in subsequent phases. The interface design proved intuitive, resource management is straightforward, and the hybrid pdfcpu/ImageMagick approach provides optimal performance and quality.

The implementation validated that:
- Interface-based design enables both planned processing patterns
- Page-level abstractions provide necessary flexibility
- Resource cleanup requirements are minimal
- Error handling patterns are clear and actionable
- Manual testing tools are essential for quality verification

These primitives are ready for integration with processing patterns (Phase 2) and caching infrastructure (Phase 3), with lessons learned documented for extraction into the go-agents-document-context library.
