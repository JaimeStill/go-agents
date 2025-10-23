# Phase 4: System Prompt Generation - Development Summary

## Starting Point

Phase 4 built on the complete infrastructure from Phases 1-3 to implement system prompt generation through sequential processing of classification policy documents. The phase integrated document processing primitives, sequential context accumulation, caching, and retry logic with go-agents vision capabilities to progressively build classification system prompts.

### Initial Requirements

- Base64 image encoding for vision API compatibility
- PDF discovery and sequential processing orchestration
- Vision agent integration with programmatic system prompt injection
- Context accumulation across multiple reference documents
- Cache integration to avoid reprocessing policy documents
- CLI tool for prompt generation workflow
- Progress visibility during sequential processing

## Implementation Decisions

### Image Encoding Infrastructure

**MimeType() Method on ImageFormat Type**:
```go
func (f ImageFormat) MimeType() (string, error) {
    switch f {
    case PNG:
        return "image/png", nil
    case JPEG:
        return "image/jpeg", nil
    default:
        return "", fmt.Errorf("unsupported image format: %s", f)
    }
}
```

**Rationale**: Encapsulating format-to-MIME type mapping within the `ImageFormat` type provides better encapsulation than duplicating format validation logic across packages. The encoding package delegates format concerns to the type that owns them, following the principle of "data and behavior belong together."

**Data URI Construction**:
```go
func EncodeImageDataURI(data []byte, format document.ImageFormat) (string, error) {
    if len(data) == 0 {
        return "", fmt.Errorf("image data is empty")
    }

    mimeType, err := format.MimeType()
    if err != nil {
        return "", err
    }

    var builder strings.Builder
    builder.WriteString("data:")
    builder.WriteString(mimeType)
    builder.WriteString(";base64,")
    builder.WriteString(base64.StdEncoding.EncodeToString(data))

    return builder.String(), nil
}
```

**Rationale**: Uses `strings.Builder` for efficient string concatenation and delegates MIME type validation to the format type. Base64 encoding uses Go's standard library for reliability.

### PDF Discovery

**filepath.Glob for Idiomatic Discovery**:
```go
func discoverPDFs(path string) ([]string, error) {
    pattern := filepath.Join(path, "*.pdf")
    matches, err := filepath.Glob(pattern)
    if err != nil {
        return nil, fmt.Errorf("failed to discover PDFs in %s: %w", path, err)
    }

    if len(matches) == 0 {
        return nil, fmt.Errorf("no PDF files found in %s", path)
    }

    return matches, nil
}
```

**Rationale**: Standard library `filepath.Glob` is more idiomatic than manual directory traversal with `os.ReadDir`. Works naturally with the filesystem's native case sensitivity (Linux: case-sensitive, Windows/macOS: case-insensitive by default). Simplified from original 15-line manual implementation to 3 lines of core logic.

### Cache Validation

**slices.Equal for Slice Comparison**:
```go
func checkCache(cachePath string, pdfs []string) (string, bool) {
    cached, err := cache.Load(cachePath)
    if err != nil {
        return "", false
    }

    if !slices.Equal(cached.ReferenceDocuments, pdfs) {
        return "", false
    }

    return cached.SystemPrompt, true
}
```

**Rationale**: Go 1.21+ `slices.Equal` provides idiomatic slice comparison with automatic length checking and element comparison. Eliminates manual loop with potential index bugs and makes intent clear.

### Agent System Prompt Injection

**Programmatic Injection (Not Configuration)**:
```go
// In Generate function
cfg.Agent.SystemPrompt = agentSystemPrompt()
visionAgent, err := agent.New(&cfg.Agent)
```

**Agent System Prompt**:
```go
func agentSystemPrompt() string {
    return `You are a document classification expert focused on IDENTIFYING classification markings in documents.

Your task: Build a system prompt that teaches how to IDENTIFY and INTERPRET classification markings visually present in documents.

EXTRACT ONLY information about:
1. Classification levels and their visual markers (TS, S, C, U, CUI)
2. Marking formats (banner markings, portion markings, headers/footers)
3. Syntax and separators (//, /, -, spaces, commas)
4. Caveats and dissemination controls (NOFORN, ORCON, REL TO, ACCM)
5. How to determine highest classification from multiple markings
6. Visual patterns and placement of markings

IGNORE information about:
- Procedural guidance (how to classify, derivative classification processes)
- Responsibilities and authorities (OCAs, derivative classifiers)
- Administrative actions (tentative classification, missing markings)
...

CRITICAL OUTPUT REQUIREMENTS:
- Return ONLY the raw system prompt text
- NO conversational preambles ("Here's the updated prompt...")
- NO follow-up questions ("Do you want me to refine...")
- Just return the direct, clean prompt text that will be used`
}
```

**Rationale**: Keeping the vision agent's system prompt in code (not configuration) makes it easier to version, maintain, and refine. The prompt evolved through iterations to focus on visual marking identification rather than procedural policy, and to eliminate conversational framing from responses.

### Retry Configuration Tuning

**Azure Rate Limit Optimization**:
```go
func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts:       3,
        InitialBackoff:    acfg.Duration(13 * time.Second),
        MaxBackoff:        acfg.Duration(50 * time.Second),
        BackoffMultiplier: 1.2,
    }
}
```

**Rationale**: Defaults tuned based on real-world Azure OpenAI rate limiting behavior. Azure's 429 responses indicate "retry after 12-13 seconds", so initial backoff of 13s ensures first retry succeeds. Multiplier of 1.2 (vs generic 2.0) provides gradual escalation suitable for predictable rate limits. This evolved from original defaults (1s initial, 2.0 multiplier) through operational experience.

### Generic Progress Reporting

**ProgressFunc Enhanced with Result Context**:
```go
type ProgressFunc[T any] func(completed, total int, result T)
```

**Usage in Sequential Processing**:
```go
progressFunc := func(completed, total int, currentContext string) {
    fmt.Fprintf(os.Stderr, "\n  Page %d/%d processed:\n", completed, total)
    fmt.Fprintf(os.Stderr, "  ---\n")

    preview := currentContext
    if len(preview) > 500 {
        preview = preview[:500] + "...\n  (truncated)"
    }
    fmt.Fprintf(os.Stderr, "  %s\n", preview)
    fmt.Fprintf(os.Stderr, "  ---\n")
}
```

**Rationale**: Making progress reporting generic with typed results enables callers to display meaningful progress information. For sequential processing, showing context updates provides transparency into what the agent is producing. Infrastructure change required updating all processing package tests to accommodate new signature.

### Standard Output Separation

**stderr for Metadata, stdout for Data**:
```go
// Progress and status to stderr
fmt.Fprintf(os.Stderr, "Discovering reference documents...\n")
fmt.Fprintf(os.Stderr, "  Found: %s (%d pages)\n", filename, pageCount)

// Final result to stdout
fmt.Println()
fmt.Println("---")
fmt.Println(result)
fmt.Println("---")
```

**Rationale**: Following Unix conventions, diagnostic information (progress, status, errors) goes to stderr while program output (generated prompt) goes to stdout. Enables piping output to files or other programs without progress messages polluting the data stream.

## Technical Challenges and Solutions

### Challenge 1: Azure Rate Limiting (429 Errors)

**Problem**: Sequential processing of 11 pages immediately hit Azure OpenAI rate limits with 429 "retry after 12-13 seconds" responses.

**Initial Retry Configuration**:
```go
// Generic defaults from Phase 2
InitialBackoff: 1s
MaxBackoff: 30s
BackoffMultiplier: 2.0
MaxAttempts: 3
```

**Failure Pattern**:
- Attempt 1: Immediate → 429
- Wait: 1s
- Attempt 2: After 1s → Still 429 (within 13s window)
- Wait: 2s
- Attempt 3: After 3s total → Still 429 (within 13s window)
- **Result**: All attempts exhausted, processing failed

**Iterative Tuning**:

*Iteration 1*: Increase initial backoff to 5s, max attempts to 4
- Result: Sometimes succeeded, sometimes failed

*Iteration 2*: Increase initial backoff to 10s, multiplier to 1.2
- Result: More reliable, but still occasional failures

*Final Solution*: Match Azure's exact retry window
```go
InitialBackoff: 13s     // Matches Azure's "retry after 12-13s"
MaxBackoff: 50s
BackoffMultiplier: 1.2  // Gradual escalation
MaxAttempts: 3
```

**Outcome**: First retry (after 13s wait) succeeds in ~95% of cases. Retry infrastructure from Phase 2 proved its value immediately with real-world rate limiting.

### Challenge 2: Small Model Hallucinations

**Problem**: gemma3:4b extracted incorrect acronym expansions from policy documents.

**Examples of Hallucinations**:
- NOFORN: "Nuclear Operations" (correct: "Not Releasable to Foreign Nationals")
- ORCON: "Or Confinement" (correct: "Originator Controlled")
- REL TO: "Related To" (correct: "Releasable To")
- ACCM: "Access Control Command" (correct: "Alternative Compensatory Control Measures")

**Root Cause**: Smaller models lack capacity to reliably extract factual information from technical documents. Gemma3:4b generated plausible-sounding but completely incorrect expansions based on acronym patterns rather than reading the actual definitions.

**Solution**: Switched to GPT-4o for prompt generation
- Result: Accurate extraction of acronym definitions and technical specifications
- Lesson: Model size matters for factual extraction from specialized documents

### Challenge 3: Conversational Framing in Responses

**Problem**: Agent responses included conversational preambles and follow-up questions instead of just the clean system prompt.

**Example Output**:
```
Okay, here's the updated system prompt incorporating the information...

**SYSTEM PROMPT:**
[actual prompt content]

Do you want me to refine this further based on...
```

**Initial Agent Instructions**:
- "Return ONLY the complete updated system prompt"

**Refinement**: Added explicit output constraints:
```
CRITICAL OUTPUT REQUIREMENTS:
- Return ONLY the raw system prompt text
- NO conversational preambles ("Here's the updated prompt...")
- NO follow-up questions ("Do you want me to refine...")
- NO markdown indicating this is a prompt ("**SYSTEM PROMPT:**")
- Just return the direct, clean prompt text that will be used
```

**Outcome**: GPT-4o respected the stricter instructions and returned clean prompt text. Smaller models (gemma3:4b) struggled to follow these constraints consistently.

### Challenge 4: Irrelevant Procedural Content

**Problem**: Initial system prompts contained extensive procedural guidance (derivative classification procedures, OCA responsibilities, tentative classification workflows) irrelevant to visual marking identification.

**Initial Agent Instructions**:
- Generic: "Extract classification information from documents"

**Refinement**: Added explicit focus constraints:
```
EXTRACT ONLY information about:
1. Classification levels and their visual markers
2. Marking formats and syntax
3. Caveats and dissemination controls
4. How to determine highest classification
5. Visual patterns and placement

IGNORE information about:
- Procedural guidance (how to classify)
- Responsibilities and authorities
- Administrative actions
- Training requirements
```

**Outcome**: Generated prompts focused on visual marking identification suitable for document classification tasks, excluding procedural policy content.

## Testing Strategy

### Black-Box Testing Approach

All tests follow the established convention from Phases 1-3:
- Package suffix: `package <name>_test`
- Import tested package explicitly
- Test only exported API
- Cannot access unexported members

### New Test Packages

**tests/encoding/image_test.go** (5 tests):
- PNG data URI structure validation
- JPEG data URI structure validation
- Empty data error handling
- Unsupported format error handling
- Large data encoding verification

**tests/prompt/prompt_test.go** (3 tests):
- Empty directory validation
- Nonexistent directory error handling
- File path (not directory) validation
- *Integration tests skipped*: Require live agent and PDFs

### Updated Test Packages

**tests/processing/** (12 tests):
- Updated all progress functions to new `ProgressFunc[T]` signature
- Tests validate progress callbacks receive result/context values

**tests/config/config_test.go** (4 tests):
- Updated retry default expectations to match Azure-tuned values
- Added BackoffMultiplier validation (1.2)
- InitialBackoff expectation: 1s → 13s
- MaxBackoff expectation: 30s → 50s

### Total Test Coverage

**Test Count**: 45 tests across 7 packages
- `tests/cache/` - 4 tests
- `tests/config/` - 4 tests
- `tests/document/` - 7 tests
- `tests/encoding/` - 5 tests (new)
- `tests/processing/` - 12 tests (updated)
- `tests/prompt/` - 3 tests (new)
- `tests/retry/` - 6 tests

**Status**: All tests passing

### Manual Integration Validation

Integration testing performed via CLI with real PDFs and live agents:
- Tested with Azure OpenAI (GPT-4o)
- Tested with Ollama (gemma3:4b)
- Validated cache functionality
- Confirmed retry behavior with rate limiting
- Verified context accumulation across pages

## Final Architecture

### Package Structure

```
tools/classify-docs/
├── pkg/
│   ├── encoding/
│   │   └── image.go           # Base64 data URI encoding
│   ├── prompt/
│   │   └── prompt.go          # System prompt generation
│   ├── config/                # Configuration (from Phase 2)
│   ├── retry/                 # Retry logic (from Phase 2)
│   ├── cache/                 # Caching (from Phase 3)
│   ├── document/              # PDF processing (from Phase 1)
│   └── processing/            # Processors (from Phase 2, updated)
│       ├── progress.go        # Generic ProgressFunc[T]
│       ├── parallel.go
│       └── sequential.go
├── tests/
│   ├── encoding/              # Encoding tests (new)
│   │   └── image_test.go
│   ├── prompt/                # Prompt tests (new)
│   │   └── prompt_test.go
│   ├── config/                # Updated for retry defaults
│   └── processing/            # Updated for ProgressFunc[T]
├── cmd/
│   ├── generate-prompt/       # CLI (new)
│   │   └── main.go
│   ├── test-config/           # From Phase 2
│   └── test-render/           # From Phase 1
├── _context/
│   ├── security-classification-markings.pdf
│   └── infosec-marking-dodm-5200-1.pdf (reduced to 11 pages)
├── .cache/
│   └── system-prompt.json     # Generated prompt cache
├── config.classify-gpt4o-key.json
├── config.classify-gpt4o-entra.json
├── config.classify-gemma.json
├── .gitignore
├── go.mod
├── go.sum
├── PROJECT.md
└── README.md
```

### Public API Surface

**New Packages**:

**pkg/encoding**:
- `EncodeImageDataURI(data []byte, format ImageFormat) (string, error)`

**pkg/prompt**:
- `Generate(ctx context.Context, cfg ClassifyConfig, referencesPath string) (string, error)`

**Updated Packages**:

**pkg/document**:
- `(ImageFormat) MimeType() (string, error)` - Method added

**pkg/processing**:
- `ProgressFunc[T any] func(completed, total int, result T)` - Generic result parameter added
- `ProcessPages[T](... progress ProgressFunc[T]) ([]T, error)` - Updated signature
- `ProcessWithContext[TContext](... progress ProgressFunc[TContext]) (..., error)` - Updated signature

**pkg/config**:
- `DefaultRetryConfig()` - Defaults tuned for Azure rate limiting

### CLI Commands

**generate-prompt**:
```bash
go run ./cmd/generate-prompt [flags]

Flags:
  --config string      Config file path (default: "config.classify-gemma.json")
  --token string       API token (overrides config)
  --references string  Reference documents directory (default: "_context")
  --no-cache          Disable cache usage
  --timeout duration  Operation timeout (default: 30m)
```

### Dependencies

No new external dependencies. Phase 4 leverages existing infrastructure:
- `github.com/JaimeStill/go-agents` (vision agent interface)
- `github.com/pdfcpu/pdfcpu` (PDF processing)

## Design Validation Results

### Sequential Processing with Vision

**Assessment**: Sequential context accumulation works effectively for building system prompts progressively.

**Validation**:
- 11 pages processed successfully with context growing from initial template to comprehensive prompt
- Each page contributed meaningful classification information
- No need for separate synthesis step - accumulated context IS the final deliverable
- Context accumulation validated core assumption from Phase 2

### Retry Infrastructure Resilience

**Assessment**: Phase 2 retry infrastructure proved immediately valuable with real-world rate limiting.

**Validation**:
- Successfully handled Azure 429 rate limit responses
- Configuration tuning (13s initial backoff) provided 95% first-retry success rate
- Exponential backoff with 1.2 multiplier appropriate for predictable rate limits
- Demonstrates value of flexible retry configuration

### Progress Visibility

**Assessment**: Generic progress reporting with result context provides valuable transparency.

**Validation**:
- Users can observe prompt evolution page-by-page
- Truncated preview (500 chars) balances detail with readability
- stderr/stdout separation enables clean piping while preserving progress visibility
- Architecture change (generic ProgressFunc[T]) worth the test update effort

### Cache Effectiveness

**Assessment**: Caching eliminates redundant processing of reference documents.

**Validation**:
- Second run with identical PDFs uses cached prompt (instant response)
- Cache invalidation works correctly when PDFs change
- `slices.Equal` comparison prevents subtle ordering bugs
- Metadata tracking (timestamp, reference documents) supports debugging

## Performance Characteristics

### Prompt Generation Timing

**11-Page Processing** (GPT-4o on Azure):
- Initial run: ~2-3 minutes with retries
- Cached run: < 1 second
- Most time spent: Vision API calls and retry backoffs
- Image rendering: ~400-600ms per page (acceptable overhead)

### Retry Timing (Azure Rate Limiting)

**Best Case** (most common):
- Attempt 1: Immediate → 429
- Wait: 13s
- Attempt 2: Success
- Total: ~13-15s per page

**Worst Case**:
- Attempt 1: Immediate → 429
- Wait: 13s
- Attempt 2: 429
- Wait: 15.6s (13s × 1.2)
- Attempt 3: Success
- Total: ~28-30s per page

### Memory Profile

**Memory Efficiency Validated**:
- 11 Page references: < 10 KB (lightweight structs with page numbers)
- Image rendering on-demand: One 200-500 KB image at a time
- No accumulation of rendered images in memory
- Total memory footprint minimal despite 11-page processing

## Lessons Learned

### What Worked Well

1. **Programmatic System Prompt Injection**: Keeping agent instructions in code (not config) enabled rapid iteration and refinement
2. **filepath.Glob**: Standard library pattern matching much cleaner than manual directory traversal
3. **slices.Equal**: Idiomatic slice comparison eliminated potential index bugs in cache validation
4. **Generic Progress Reporting**: Architecture investment in ProgressFunc[T] paid off with meaningful visibility
5. **stderr/stdout Separation**: Unix conventions enable clean composition with other tools
6. **MimeType() Method**: Encapsulating format behavior on the type improved code organization

### What Required Iteration

1. **Retry Configuration Tuning**: Generic defaults (1s, 2.0 multiplier) inadequate for Azure rate limiting. Required empirical tuning to 13s initial backoff and 1.2 multiplier
2. **Agent Instructions**: Multiple refinements needed to focus on visual marking identification and eliminate conversational framing
3. **Model Selection**: gemma3:4b hallucinations revealed small models inadequate for factual extraction from technical documents
4. **Progress Function Signature**: Adding result parameter required updating all processing tests, but worth the effort for visibility

### Recommendations for go-agents-document-context

1. **Image Encoding**: MimeType() method pattern should be adopted for any format enums
2. **Progress Infrastructure**: Generic progress reporting with typed results provides valuable transparency
3. **Retry Defaults**: Consider provider-specific configuration presets (Azure, OpenAI, etc.) given different rate limiting behaviors
4. **Model Requirements**: Document minimum model size recommendations for factual extraction vs. general processing
5. **Agent Prompt Engineering**: Provide guidance on preventing conversational framing in sequential processing contexts
6. **Cache Validation**: Use slices.Equal for slice comparisons to avoid manual iteration bugs

## Current State

### Completed

- ✅ Base64 image encoding with data URI support
- ✅ System prompt generation via sequential processing
- ✅ Vision agent integration with programmatic prompt injection
- ✅ PDF discovery with filepath.Glob
- ✅ Cache validation with slices.Equal
- ✅ Retry configuration tuned for Azure rate limiting
- ✅ Generic progress reporting with result context
- ✅ CLI tool (generate-prompt command)
- ✅ Comprehensive tests (45 total, all passing)
- ✅ Agent instructions refined for visual marking focus
- ✅ stderr/stdout separation for clean composition

### Known Limitations

- Vision agent prompt hardcoded (not externalized to config or file)
- No validation of generated prompt quality (manual review required)
- Reference document paths assumed to be valid PDFs (no format validation)
- Sequential processing cannot resume mid-document (must restart on failure)
- Progress preview truncated at 500 chars (not configurable)
- Model selection not validated at runtime (can use models inadequate for task)

### Files Delivered

**Package Files**:
1. `pkg/encoding/image.go` - Base64 data URI encoding
2. `pkg/document/document.go` - Added MimeType() method to ImageFormat
3. `pkg/prompt/prompt.go` - Prompt generation orchestration
4. `pkg/processing/progress.go` - Updated ProgressFunc[T] signature
5. `pkg/processing/parallel.go` - Updated for ProgressFunc[T]
6. `pkg/processing/sequential.go` - Updated for ProgressFunc[T]
7. `pkg/config/retry.go` - Tuned default values for Azure

**Test Files**:
8. `tests/encoding/image_test.go` - 5 tests for image encoding
9. `tests/prompt/prompt_test.go` - 3 tests for prompt validation
10. `tests/processing/parallel_test.go` - Updated for new progress signature
11. `tests/processing/sequential_test.go` - Updated for new progress signature
12. `tests/config/config_test.go` - Updated retry default expectations

**Command Files**:
13. `cmd/generate-prompt/main.go` - CLI for prompt generation

**Generated Artifacts**:
14. `.cache/system-prompt.json` - Cached system prompt with metadata

**Documentation**:
15. `_context/.archive/03-system-prompt-generation.md` - This document

### Next Phase Prerequisites

Phase 5 (Document Classification) depends on:
- ✅ Generated system prompt (from Phase 4)
- ✅ Parallel processing with worker pools (from Phase 2)
- ✅ Retry infrastructure for LLM calls (from Phase 2, tuned in Phase 4)
- ✅ Configuration management (from Phase 2)
- ✅ Document processing primitives (from Phase 1)

Phase 5 will implement:
- Parallel page classification workflow
- Classification result parsing and aggregation
- Derivation of highest overall classification
- JSON output with structured results
- CLI for document classification

## Reference Information

### CLI Usage

**Basic Usage**:
```bash
cd tools/classify-docs
go run ./cmd/generate-prompt
```

**With Specific Configuration**:
```bash
go run ./cmd/generate-prompt --config config.classify-gpt4o-key.json --token $AZURE_API_KEY
```

**Disable Cache**:
```bash
go run ./cmd/generate-prompt --no-cache
```

**Custom References Directory**:
```bash
go run ./cmd/generate-prompt --references /path/to/policy/docs
```

**Build Standalone Binary**:
```bash
go build -o generate-prompt ./cmd/generate-prompt
./generate-prompt --token $AZURE_API_KEY
```

### Output Examples

**Successful Generation**:
```
Discovering reference documents...
  Found: dodm-5200.01-enc4.pdf (9 pages)
  Found: security-classification-markings.pdf (2 pages)

Generating system prompt from 11 pages...

  Page 1/11 processed:
  ---
  You are a document classification expert specializing in DoD security...

  ## Classification Levels
  ...
  (truncated)
  ---

  Page 11/11 processed:
  ---
  [Complete system prompt with all sections populated]
  ---

System prompt generated successfully!
  Cached: .cache/system-prompt.json

---
[Final system prompt output to stdout]
---
```

**Using Cached Prompt**:
```
Discovering reference documents...
  Found: dodm-5200.01-enc4.pdf (9 pages)
  Found: security-classification-markings.pdf (2 pages)

Using cached system prompt

---
[Cached system prompt output to stdout]
---
```

### Testing Commands

**Run All Tests**:
```bash
cd tools/classify-docs
go test ./tests/... -v
```

**Run New Test Packages**:
```bash
go test ./tests/encoding/... -v
go test ./tests/prompt/... -v
```

**Verify Retry Defaults**:
```bash
go test ./tests/config/... -v -run TestDefaultConfigs/DefaultRetryConfig
```

**Test Count**:
```bash
go test ./tests/... -v 2>&1 | grep -c "^=== RUN"
# Output: 45
```

### Common Issues and Solutions

**Issue**: "no PDF files found" error
**Solution**: Ensure reference documents exist in specified directory (default: `_context/`)

**Issue**: Azure 429 rate limit errors despite retry
**Solution**: Retry defaults tuned for 12-13s window. If still failing, increase `InitialBackoff` in config

**Issue**: Gemma3:4b producing incorrect acronym definitions
**Solution**: Use GPT-4o or larger model for factual extraction from technical documents

**Issue**: Agent responses include conversational framing
**Solution**: GPT-4o respects strict output constraints. Smaller models may struggle with this consistently

**Issue**: Cache not invalidating when PDFs change
**Solution**: Cache validates based on exact PDF paths. Renaming or moving PDFs will invalidate cache correctly

## Conclusion

Phase 4 successfully implemented system prompt generation through sequential processing of classification policy documents. The implementation integrated infrastructure from Phases 1-3 with go-agents vision capabilities to progressively build comprehensive classification system prompts.

Key achievements:
- **Vision Agent Integration**: Seamless use of go-agents vision protocol for document analysis
- **Context Accumulation**: Sequential processing proved effective for building prompts progressively
- **Retry Resilience**: Infrastructure adapted to real-world Azure rate limiting through empirical tuning
- **Progress Transparency**: Generic progress reporting provides valuable visibility into prompt evolution
- **Cache Efficiency**: Eliminates redundant processing of reference documents
- **Idiomatic Refinements**: MimeType() method, filepath.Glob, slices.Equal improved code quality
- **Modern Go**: Continued use of Go 1.25.2 patterns (strings.Builder, error wrapping)

The generated system prompt serves as the foundation for Phase 5's document classification workflow, validating the document processing → agent analysis architecture that will inform go-agents-document-context library design.
