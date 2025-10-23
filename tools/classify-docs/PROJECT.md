# Document Classification POC

## Project Overview

The `classify-docs` tool is a proof-of-concept implementation that validates the document processing → agent analysis architecture. This tool serves dual purposes:

1. **Immediate Value**: Classify security markings in DoD documents using go-agents vision capabilities
2. **Architecture Validation**: Prototype design patterns for the future go-agents-document-context library

### Key Principle

This POC demonstrates that **traditional programming techniques (Go) should prepare and optimize contextual artifacts before intelligent processing by agents**. Document preprocessing is not optional—it's essential for handling documents that LLMs cannot process natively.

## Problem Context

### Azure OpenAI Vision API Limitations

Unlike OpenAI's direct API, Azure OpenAI does **not** support native PDF input. Documents must be:
1. Extracted page-by-page
2. Converted to images (PNG or JPEG)
3. Sent to vision API for analysis

This limitation validates the need for robust document processing infrastructure in Go.

### Classification Requirements

DoD security classification analysis requires:
- **Multi-classification handling**: Documents often contain multiple classification markings on the same page
- **Comprehensive marking detection**: Banner markings, portion markings, header/footer markings
- **Policy-based derivation**: Highest classification governs, with proper caveat handling
- **Page-level granularity**: Independent analysis of each page with aggregated results

### System Prompt Generation Requirements

Generating classification system prompts from policy documents requires:
- **Sequential context accumulation**: Build understanding progressively across document pages
- **Manageable token usage**: Process one page at a time with accumulated context
- **Progressive refinement**: Each page refines the system prompt based on new information
- **Focused processing**: Small context windows where agents excel

## Architecture Layers

The tool is organized into three architectural layers that will inform go-agents-document-context design:

### Layer 1: Document Processing Primitives (Reusable)

**Purpose**: Low-level, format-agnostic document operations

**Interfaces**:
```go
type Document interface {
    PageCount() int
    ExtractPage(pageNum int) (Page, error)
    ExtractAllPages() ([]Page, error)
    Close() error
}

type Page interface {
    Number() int
    ToImage(opts ImageOptions) ([]byte, error)  // For vision processing
}
```

**Implementation**:
- PDF processing using `pdfcpu` (pure Go, no CGo)
- Page extraction and image conversion
- Resource cleanup and lifecycle management

**Design Questions Answered** (Phase 1):
- ✅ Interfaces are sufficient for both parallel and sequential processing
- ✅ Document-level cleanup necessary; Page-level cleanup (`Close()`) unnecessary
- ✅ pdfcpu provides fast structure operations; ImageMagick handles rendering (~400-600ms per page at 150 DPI)

### Layer 2: Processing Patterns (Orchestration)

**Purpose**: Document processing workflows

**Patterns**:

**Parallel Processing**: Independent page analysis
```go
type ParallelProcessor struct {
    Workers int  // Auto-detected: runtime.NumCPU() * 2, capped at 16
}
```
- Pages processed concurrently
- Results aggregated maintaining order
- Fail-fast on errors
- **Use Case**: Document classification (each page analyzed independently)

**Sequential Processing**: Context accumulation
```go
type SequentialProcessor struct {
    // Minimal configuration
}
```
- Pages processed in order
- Context accumulated across pages
- Each page refines accumulated context
- **Use Case**: System prompt generation (progressive understanding)

**Key Insight**: The accumulated context from sequential processing **IS** the final deliverable. No separate synthesis step needed.

**Design Questions to Answer**:
- Does auto-detection of worker count work well?
- How to handle partial failures gracefully?
- What's the performance gain from parallelization?
- Is context accumulation effective for prompt generation?

### Layer 3: Use-Case Implementations (Application-Specific)

**Purpose**: Domain-specific applications

**Implementations**:
- **System Prompt Generation**: Sequential processing with context accumulation
- **Document Classification**: Parallel processing for independent page analysis

**Design Questions to Answer**:
- Can the same primitives serve both use cases effectively?
- What additional abstractions are needed?
- How to structure application-specific logic?

## Supporting Infrastructure

### Image Caching

**Purpose**: Avoid reprocessing documents

**Strategy**:
- Cache key: Hash of (file path + modification time)
- Structure: `<cache-dir>/<doc-hash>/page-NNN.png`
- Validation: Compare source mod time vs cache time
- Configuration: `--cache-dir` flag, `--no-cache` to disable

**Benefits**:
- Avoid re-extracting pages from PDFs
- Reuse images across multiple operations
- Significant performance improvement on repeated operations

**Design Questions to Answer**:
- Does caching provide meaningful performance improvement?
- What's the appropriate cache eviction strategy?
- Should cache be part of core library or application concern?

### Configuration Management

**Agent Configuration**: Based on `tools/prompt-agent/config.gpt-4o.json`
- Vision capability with high detail
- Azure OpenAI provider
- API key injection via `--token` flag

**Image Options**:
- Format: PNG (lossless, higher quality) - configurable
- DPI: 150 (balance quality/size) - configurable
- Quality: N/A for PNG, 80-100 for JPEG

## Phase Development Plan

### Phase 1: Document Processing Primitives ✅

**Status**: Complete

**Development Summary**: `_context/.archive/01-document-processing-primitives.md`

**Objectives**:
- ✅ Define core `Document` and `Page` interfaces
- ✅ Implement PDF processing using pdfcpu
- ✅ Implement page-to-image conversion (PNG and JPEG)
- ✅ Validate resource management patterns

**Deliverables**:
- ✅ `document/document.go` - Core interfaces and types
- ✅ `document/pdf.go` - PDF implementation
- ✅ `tests/document/pdf_test.go` - Unit tests for page extraction and image conversion
- ✅ `cmd/test-render/main.go` - Manual testing tool

**Success Criteria**:
- ✅ Successfully extract pages from multi-page PDF
- ✅ Convert pages to PNG and JPEG images with configurable DPI
- ✅ Proper resource cleanup (Document-level cleanup, Page-level cleanup unnecessary)
- ✅ Clean error handling and propagation
- ✅ Manual testing tool for visual quality verification

### Phase 2: Processing Infrastructure ✅

**Status**: Complete

**Development Summary**: `_context/.archive/02-processing-infrastructure.md`

**Objectives**:
- ✅ Implement generic parallel processing with worker pools
- ✅ Implement generic sequential processing with context accumulation
- ✅ Auto-detect optimal worker count for parallel processing
- ✅ Fail-fast error handling with context coordination
- ✅ Result aggregation maintaining page order
- ✅ Retry infrastructure with exponential backoff
- ✅ Unified configuration management

**Deliverables**:
- ✅ `pkg/config/` - Configuration types and loading (cache, retry, processing)
- ✅ `pkg/retry/` - Generic retry with exponential backoff
- ✅ `pkg/processing/` - Parallel and sequential processors
- ✅ `tests/config/`, `tests/retry/`, `tests/processing/` - Comprehensive black-box tests
- ✅ `cmd/test-config/` - Configuration verification utility
- ✅ `config.classify-*.json` - Minimal configuration files (Azure, Ollama)

**Success Criteria**:
- ✅ **Parallel**: Process multiple pages concurrently with proper ordering
- ✅ **Parallel**: Graceful worker pool lifecycle management
- ✅ **Parallel**: Fail immediately on first error with context cancellation
- ✅ **Parallel**: Deadlock-free with background result collector
- ✅ **Sequential**: Process pages in order with context accumulation
- ✅ **Sequential**: Generic type parameter for any context type
- ✅ **Sequential**: Optional intermediate context capture
- ✅ **Retry**: Exponential backoff with configurable parameters
- ✅ **Retry**: Non-retryable error marking
- ✅ **Config**: Minimal configs with proper default merging
- ✅ **Config**: Pointer-based booleans for default `true` support

### Phase 3: Caching Infrastructure ✅

**Status**: Complete (merged into Phase 2)

**Development Summary**: `_context/.archive/02-processing-infrastructure.md`

**Objectives**:
- ✅ Implement system prompt caching
- ✅ JSON-based persistence with metadata
- ✅ Automatic directory creation
- ✅ Configuration integration

**Deliverables**:
- ✅ `pkg/cache/cache.go` - Cache implementation
- ✅ `tests/cache/cache_test.go` - Unit tests
- ✅ Cache configuration in unified config structure

**Success Criteria**:
- ✅ Cache prevents redundant system prompt generation
- ✅ Metadata tracking (timestamp, reference documents)
- ✅ Configurable cache path and enable/disable
- ✅ Proper default merging (enabled by default)

### Phase 4: System Prompt Generation ✅

**Status**: Complete

**Development Summary**: `_context/.archive/03-system-prompt-generation.md`

**Objectives**:
- ✅ Process classification guide and policy PDFs sequentially
- ✅ Use context accumulation to build system prompt progressively
- ✅ Each page refines the system prompt with new information
- ✅ Final accumulated context is the complete system prompt

**Deliverables**:
- ✅ `pkg/prompt/prompt.go` - Prompt generation logic using sequential processor
- ✅ `pkg/encoding/image.go` - Base64 data URI encoding for vision API
- ✅ `.cache/system-prompt.json` - Generated classification prompt with metadata
- ✅ `cmd/generate-prompt/main.go` - CLI tool for prompt generation
- ✅ `tests/prompt/prompt_test.go` - Unit tests for prompt generation
- ✅ `tests/encoding/image_test.go` - Unit tests for image encoding

**Success Criteria**:
- ✅ Successfully process policy documents page-by-page
- ✅ Context accumulates correctly across pages
- ✅ Final prompt captures: levels, formats, caveats, derivation rules, edge cases
- ✅ Usable for classification tasks without further refinement
- ✅ Retry infrastructure handles Azure rate limiting (13s initial backoff, 1.2 multiplier)
- ✅ Generic ProgressFunc provides visibility into context accumulation

**Workflow**:
```
Initial template → Page 1 → Updated v1 → Page 2 → Updated v2 → ... → Final prompt
```

### Phase 5: Document Classification ✅

**Status**: Complete

**Development Summary**: `_context/.archive/05-document-classification.md` (TBD)

**Objectives**:
- ✅ Implement per-page classification with context accumulation
- ✅ Design comprehensive classification prompt with self-check verification
- ✅ Implement conservative confidence scoring (HIGH/MEDIUM/LOW)
- ✅ Handle spatially separated and faded classification components
- ✅ Optimize for o4-mini visual reasoning model
- ✅ Validate accuracy on 27-document test set

**Deliverables**:
- ✅ `pkg/classify/document.go` - Classification logic with sequential processing
- ✅ `pkg/classify/parse.go` - JSON response parsing
- ✅ `cmd/classify/main.go` - CLI tool for document classification
- ✅ `config.classify-o4-mini.json` - Optimized configuration (300 DPI, reasoning_effort: high)
- ✅ `.cache/system-prompt.json` - Generated system prompt with spatial separation guidance
- ✅ `classification-results.json` - Final test results (26/27 documents, 96.3% accuracy)

**Success Criteria**:
- ✅ **Accuracy**: Achieved 96.3% accuracy (26/27 documents correct)
- ✅ **Confidence Distribution**: 22 HIGH (81.5%), 5 MEDIUM (18.5%), 0 LOW
- ✅ **Conservative Scoring**: Suspicion-based confidence for classified documents without visible caveats
- ✅ **Faded Marking Detection**: Optimized settings (300 DPI, reasoning_effort: high) for low-contrast stamps
- ✅ **JSON Output**: Structured results with classification, confidence, markings, and rationale
- ✅ **Context Accumulation**: Sequential processing maintains state across pages

**Key Implementation Details**:
- Changed from planned parallel processing to sequential processing for better context accumulation
- Conservative confidence logic: If SECRET/CONFIDENTIAL/TOP SECRET with no visible caveats, assign MEDIUM confidence (assumes potential fading)
- Spatial separation handling: Combines classification components found anywhere on page (e.g., "SECRET" in header + "NOFORN" in footer)
- Self-check verification: Model questions its findings before assigning confidence
- Optimized for o4-mini: 300 DPI rendering + reasoning_effort: "high" parameter

**Results Analysis**:
- 1 classification error: Document 19 missing NOFORN caveat (correctly flagged with MEDIUM confidence)
- 4 false-positive MEDIUM flags: Documents 17, 23, 8, 24 (legitimately no caveats, but flagged for review)
- Trade-off accepted: Better to over-flag for human review than miss actual errors in security classification

### Phase 6: Testing & Validation

**Implementation Guide**: `phase-6-guide.md` (future)

**Objectives**:
- Test system prompt generation with guide and policy documents
- Test classification with real 27-page classified document
- Validate classification accuracy
- Measure performance metrics
- Document lessons learned

**Deliverables**:
- `system-prompt.txt` - Generated from policy documents
- `classification-results.json` - Classification test results
- Performance analysis document
- Architecture recommendations for go-agents-document-context

**Success Criteria**:
- Generated system prompt is comprehensive and accurate
- Classification results are accurate for sample documents
- Acceptable performance (time, token usage)
- Clear lessons learned documented
- Validated architecture patterns for both processing patterns

## Output Structure

### Page Classification

```json
{
  "page": 1,
  "classification": "SECRET",
  "confidence": "high",
  "markings_found": [
    "SECRET//NOFORN",
    "(S)",
    "(U)"
  ],
  "classification_rationale": "Page contains SECRET banner marking and multiple portion markings. Highest classification found is SECRET with NOFORN caveat. Overall page classification: SECRET per DoD Manual 5200.1."
}
```

**Fields**:
- `page`: Page number
- `classification`: Overall page classification (UNCLASSIFIED, CONFIDENTIAL, SECRET, TOP SECRET)
- `confidence`: Agent's confidence level (high, medium, low)
- `markings_found`: **Unique** markings encountered on page
- `classification_rationale`: Explanation of classification determination

### Document Analysis

```json
{
  "document_path": "/path/to/marked-documents.pdf",
  "total_pages": 27,
  "highest_classification": "SECRET",
  "processing_time": "2m15s",
  "page_results": [
    { /* PageClassification */ },
    { /* PageClassification */ }
  ]
}
```

**Fields**:
- `document_path`: Source document path
- `total_pages`: Total pages processed
- `highest_classification`: Highest classification found across all pages
- `processing_time`: Total processing duration
- `page_results`: Array of per-page classifications

## CLI Interface

### Generate System Prompt

```bash
go run . generate-prompt \
  --token $AZURE_API_KEY \
  [--cache-dir ~/.cache/classify-docs] \
  [--config config.classification.json]
```

**Note**: Guide and policy paths are configured in code or config file (in `_context/` directory)

### Classify Document

```bash
go run . classify \
  --token $AZURE_API_KEY \
  --input /path/to/marked-documents.pdf \
  --output classification-results.json \
  [--cache-dir ~/.cache/classify-docs] \
  [--config config.classification.json] \
  [--workers 8] \
  [--no-cache]
```

## Design Validation Goals

This POC will answer critical questions for go-agents-document-context:

### Interface Design
- Are `Document`/`Page` interfaces sufficient?
- What methods are missing or unnecessary?
- How to handle format-specific features?

### Resource Management
- How to manage memory for large documents?
- When to keep pages in memory vs. process incrementally?
- What cleanup patterns work best?

### Image Processing
- PNG vs JPEG quality/size tradeoffs?
- Optimal DPI for vision processing?
- How to handle Azure's 20MB image size limit?

### Processing Patterns
- **Parallel**: What worker pool size is optimal?
- **Parallel**: How to handle errors in parallel workflows?
- **Parallel**: What's the performance benefit of parallelization?
- **Sequential**: Does context accumulation work for prompt generation?
- **Sequential**: What's the token usage profile across pages?
- **Sequential**: How to structure the context update instructions?

### Error Handling
- Fail-fast vs. partial results?
- Error recovery strategies?
- How to provide actionable error messages?

### Caching Strategy
- Does caching provide meaningful value?
- What should be cached (images, results, both)?
- Cache eviction and invalidation strategies?

## Success Criteria

### Phase 1: Complete ✅
- ✅ Successfully extract and process multi-page PDFs
- ✅ Convert pages to images (PNG and JPEG) without quality loss
- ✅ Clean, idiomatic Go code with comprehensive error handling
- ✅ Efficient resource usage (Document cleanup, lightweight Page references)
- ✅ Configurable image options (format, DPI, quality)
- ✅ Interfaces validated as clean and extensible
- ✅ Code organization supports library extraction
- ✅ Manual testing tool for quality verification

### Phase 2: Complete ✅
- ✅ Generic parallel and sequential processing patterns implemented
- ✅ Worker pool with auto-detection (NumCPU * 2, capped at 16)
- ✅ Context accumulation for sequential processing
- ✅ Deadlock-free error handling with background result collection
- ✅ Retry infrastructure with exponential backoff
- ✅ Configuration management with proper default merging
- ✅ Comprehensive black-box tests (33 tests, all passing)
- ✅ Modern Go 1.25.2 patterns (WaitGroup.Go(), range over integer, etc.)

### Phase 3: Complete ✅
- ✅ System prompt caching implementation
- ✅ JSON-based persistence with metadata
- ✅ Configuration integration with pointer-based defaults
- ✅ Automatic directory creation and cleanup

### Phase 4: Complete ✅
- ✅ Generate comprehensive system prompt from policy documents using sequential processing
- ✅ Image encoding for vision API with base64 data URIs
- ✅ Retry infrastructure tuned for Azure rate limiting
- ✅ Generic progress reporting with result visibility
- ✅ CLI tool for system prompt generation
- ✅ Comprehensive black-box tests (45 tests total, all passing)

### Phase 5: Complete ✅
- ✅ Per-page document classification with sequential processing and context accumulation
- ✅ Conservative confidence scoring with suspicion-based logic for missing caveats
- ✅ Achieved 96.3% accuracy (26/27 documents) on test set
- ✅ Optimized for o4-mini visual reasoning model (300 DPI, reasoning_effort: high)
- ✅ Comprehensive classification prompt with self-check verification
- ✅ Handles spatially separated and faded classification components

### Phase 6: Planned
- Comprehensive testing across different model configurations
- Performance analysis and optimization recommendations
- Document lessons learned for go-agents-document-context library design

## Future Library Extraction

After POC completion, validated patterns will inform go-agents-document-context:

### Core Library (`go-agents-document-context`)
- Document/Page interfaces from `document/`
- PDF processor implementation
- Additional format processors (DOCX, XLSX, PPTX, images)
- Both processing patterns (parallel, sequential)
- Context optimization utilities
- Caching infrastructure (if validated as valuable)

### What Remains Application-Specific
- Classification logic and prompts
- System prompt generation instructions
- Result aggregation and formatting
- CLI interface and configuration
- Domain-specific error handling

### Open Questions for Library Design
- Should caching be part of the library or application concern?
- How to handle provider-specific constraints (Azure 20MB limit)?
- What abstractions support both vision and text extraction use cases?
- How to make format processors pluggable?
- Should sequential processing be generalized beyond context strings?

## References

### External Dependencies
- **pdfcpu**: PDF manipulation (https://github.com/pdfcpu/pdfcpu)
- **go-agents**: Agent interface library (../../)

### Related Documents
- `../../PROJECT.md` - go-agents library roadmap
- `../../ARCHITECTURE.md` - go-agents architecture
- `_context/.archive/01-document-processing-primitives.md` - Phase 1 development summary
- `_context/.archive/02-processing-infrastructure.md` - Phase 2 & 3 development summary
- `_context/.archive/03-system-prompt-generation.md` - Phase 4 development summary

### Context Documents
- `_context/security-classification-markings.pdf` - Classification guide
- `_context/infosec-marking-dodm-5200-1.pdf` - DoD Manual 5200.1

### Azure Documentation
- Azure OpenAI Vision API: https://learn.microsoft.com/en-us/azure/ai-services/openai/how-to/gpt-with-vision
- PDF Processing Sample: https://github.com/Azure-Samples/azure-openai-gpt-4-vision-pdf-extraction-sample
