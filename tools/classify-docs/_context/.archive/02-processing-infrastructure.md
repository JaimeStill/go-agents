# Phase 2 & 3: Processing Infrastructure and Caching - Development Summary

## Starting Point

Phase 2 built on Phase 1's document processing primitives to create the orchestration layer for both parallel (concurrent page processing) and sequential (context accumulation) patterns. Phase 3 (caching infrastructure) was merged into Phase 2 during implementation as it shared configuration concerns and provided natural integration points.

### Initial Requirements

- Generic processing layer supporting any result/context type
- Parallel processor with worker pools for independent page classification
- Sequential processor with context accumulation for prompt generation
- Retry logic with exponential backoff for transient failures
- System prompt caching to avoid reprocessing reference documents
- Unified configuration combining agent settings with processing parameters
- Support for omitting default values in configuration files

## Implementation Decisions

### Package Organization

**Adopted pkg/ Structure**:
```
tools/classify-docs/
├── pkg/
│   ├── config/        # All configuration types
│   ├── retry/         # Retry logic with exponential backoff
│   ├── cache/         # System prompt persistence
│   ├── document/      # Document processing (from Phase 1)
│   └── processing/    # Parallel and sequential processors
├── tests/             # Black-box tests mirroring pkg/
├── cmd/               # Command-line utilities
└── _context/          # Documentation and guides
```

**Rationale**: The pkg/ organization clearly separates package infrastructure from command utilities and tests, following Go community conventions for tools and applications.

### Generic Processing Functions

**Package-Level Generic Functions**:
```go
func ProcessPages[T any](
    ctx context.Context,
    cfg config.ParallelConfig,
    pages []document.Page,
    processor func(context.Context, document.Page) (T, error),
    progress ProgressFunc,
) ([]T, error)

func ProcessWithContext[TContext any](
    ctx context.Context,
    cfg config.SequentialConfig,
    pages []document.Page,
    initialContext TContext,
    processor ContextProcessor[TContext],
    progress ProgressFunc,
) (SequentialResult[TContext], error)
```

**Rationale**: Go methods cannot have type parameters, so package-level functions provide maximum flexibility. The generic type parameters enable type-safe processing for any result or context type without casting.

### Retry Infrastructure

**Technology Choice**: Custom implementation rather than third-party library.

**Core Design**:
```go
type Retryable[T any] func(ctx context.Context, attempt int) (T, error)

func Do[T any](
    ctx context.Context,
    cfg config.RetryConfig,
    fn Retryable[T],
) (T, error)
```

**Features**:
- Exponential backoff with configurable multiplier and max backoff
- Context coordination for fail-fast across workers
- Non-retryable error marking for permanent failures
- Attempt number passed to function for logging/tracking

**Rationale**: Simple custom implementation provides exactly what's needed without external dependencies, with context integration for coordinated cancellation.

### Configuration Architecture

**Unified Configuration Structure**:
```go
type ClassifyConfig struct {
    Agent      acfg.AgentConfig  // From go-agents
    Processing ProcessingConfig  // Tool-specific settings
}
```

**Pointer-Based Boolean for Defaults**:
```go
type CacheConfig struct {
    Enabled *bool  `json:"enabled,omitempty"`
    Path    string `json:"path,omitempty"`
}

func (c *CacheConfig) IsEnabled() bool {
    if c.Enabled == nil {
        return true // Default to enabled
    }
    return *c.Enabled
}
```

**Rationale**: JSON unmarshaling sets unspecified booleans to `false`, making it impossible to distinguish "not set" from "explicitly false". Using `*bool` with omitempty allows proper default merging - nil means "use default", non-nil means "explicit value".

### Caching Strategy

**Simple JSON-Based Persistence**:
```go
type SystemPromptCache struct {
    GeneratedAt        time.Time
    ReferenceDocuments []string
    SystemPrompt       string
}
```

**Design Decisions**:
- Human-readable JSON format for debuggability
- Metadata tracking (timestamp, source documents)
- Automatic directory creation
- Cache validity checking delegated to caller
- No dependency on config package (removed from New() signature)

**Rationale**: Keep cache package simple - it handles serialization mechanics. Calling code uses config to determine policy (enabled/disabled, path).

### Modern Go 1.25.2 Patterns

**WaitGroup.Go()** (New in Go 1.25.0):
```go
var wg sync.WaitGroup
for range workers {
    wg.Go(func() {
        // Worker logic with implicit Done()
    })
}
wg.Wait()
```

**Range Over Integer**:
```go
for range workers {  // Clean iteration without index variable
    wg.Go(func() { ... })
}
```

**Chained min()**:
```go
workers := min(min(runtime.NumCPU()*2, cfg.WorkerCap), len(pages))
```

**Deferred Channel Closure**:
```go
go func() {
    defer close(workCh)  // Always closes, even on early return
    for i, page := range pages {
        // Send work...
    }
}()
```

**Direct Context Check**:
```go
if err := ctx.Err(); err != nil {
    return result, fmt.Errorf("processing cancelled: %w", err)
}
```

## Technical Challenges and Solutions

### Challenge 1: Parallel Processing Deadlock

**Problem**: Test `TestProcessPages_Error` hung indefinitely when errors occurred during parallel processing.

**Root Cause**: Workers could exit early on context cancellation without sending results to the result channel. The collector loop waited for exactly `len(pages)` results, causing a deadlock when fewer results were sent.

**Initial Code**:
```go
// Collector waits for fixed number of results
for range pages {
    res := <-resultCh
    // Process result...
}
wg.Wait()
close(resultCh)
```

**Failed Approach**: Tried to detect cancellation in collector loop, but timing was unreliable.

**Final Solution**: Move result collection to background goroutine that drains channel until closed:
```go
// Collect results in background
var firstErr error
completed := 0
done := make(chan struct{})

go func() {
    defer close(done)
    for res := range resultCh {  // Drain until channel closed
        if res.err != nil {
            if firstErr == nil {
                firstErr = res.err
                cancel()
            }
        } else {
            results[res.index] = res.value
        }
        completed++
        if progress != nil {
            progress(completed, len(pages))
        }
    }
}()

// Wait for workers to finish
wg.Wait()
close(resultCh)  // Signal collector to finish

// Wait for result collection to complete
<-done
```

**Outcome**: Eliminated deadlock by decoupling result collection from worker completion. Collector drains all results sent before channel closes, regardless of count.

### Challenge 2: Time-Dependent Tests

**Problem**: `TestProcessPages_ContextCancellation` was flaky, sometimes failing due to timing issues.

**Initial Code**:
```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel after short delay
go func() {
    time.Sleep(50 * time.Millisecond)
    cancel()
}()

results, err := processing.ProcessPages(ctx, cfg, pages, processor, nil)
```

**Issue**: Test depended on precise timing - if processing completed too quickly or too slowly, test could fail.

**Failed Approach**: Tried adjusting sleep times, but this made tests slower and still unreliable.

**Final Solution**: Use pre-cancelled context for deterministic test:
```go
// Create an already-cancelled context
ctx, cancel := context.WithCancel(context.Background())
cancel() // Cancel immediately before processing

results, err := processing.ProcessPages(ctx, cfg, pages, processor, nil)

if err == nil {
    t.Fatal("expected error from context cancellation")
}
```

**Outcome**: Test became completely deterministic - no timing dependencies, runs instantly, never flakes.

### Challenge 3: Boolean Defaults in Configuration

**Problem**: Cache enabled defaulted to `false` instead of `true` when omitted from config files.

**Root Cause**: JSON unmarshaling assigns zero values to missing fields. For boolean fields, zero value is `false`, so:
```json
{
  "agent": { "name": "test" }
  // cache.enabled not specified
}
```

Results in `Enabled: false` after unmarshaling, which then overwrites the default `true` during merge.

**Failed Approach**: Tried conditional merge based on zero value:
```go
if source.Enabled != false {  // Doesn't work - can't distinguish explicit false
    c.Enabled = source.Enabled
}
```

**Final Solution**: Use pointer with nil check:
```go
type CacheConfig struct {
    Enabled *bool `json:"enabled,omitempty"`  // Pointer allows nil
}

func (c *CacheConfig) Merge(source *CacheConfig) {
    if source.Enabled != nil {  // Only merge if explicitly set
        c.Enabled = source.Enabled
    }
}

func (c *CacheConfig) IsEnabled() bool {
    if c.Enabled == nil {
        return true  // Default when not specified
    }
    return *c.Enabled
}
```

**Outcome**: Can now distinguish three states: not set (nil = use default true), explicitly true, explicitly false. Config files can omit `enabled` and get default `true` behavior.

## Testing Strategy

### Black-Box Testing Approach

**Package Convention**: All tests use `package <name>_test` to test public API only.

**Rationale**: Black-box testing validates library from consumer perspective, prevents coupling to internal implementation, and reduces test brittleness during refactoring.

**Example**:
```go
package config_test

import (
    "testing"
    "github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
)

func TestLoadClassifyConfig(t *testing.T) {
    // Can only access exported API
}
```

### Table-Driven Tests

**Pattern**:
```go
tests := []struct {
    name         string
    input        string
    expected     string
    expectError  bool
}{
    {
        name:     "success case",
        input:    "test",
        expected: "result",
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result, err := Function(tt.input)
        // Assertions...
    })
}
```

**Coverage**: Used extensively in config, retry, and processing tests for comprehensive scenario coverage.

### Test Organization

**Structure**:
```
tests/
├── cache/
│   └── cache_test.go           # 4 tests
├── config/
│   └── config_test.go          # 4 tests
├── document/
│   └── pdf_test.go             # 7 tests
├── processing/
│   ├── parallel_test.go        # 6 tests
│   └── sequential_test.go      # 6 tests
└── retry/
    └── retry_test.go           # 6 tests
```

**Total**: 33 tests, all passing

### Mock Pattern for Document Tests

**Implementation**:
```go
type mockPage struct {
    number int
}

func (m *mockPage) Number() int {
    return m.number
}

func (m *mockPage) ToImage(opts document.ImageOptions) ([]byte, error) {
    return []byte(fmt.Sprintf("image-%d", m.number)), nil
}
```

**Usage**: Processing tests use mock pages to avoid ImageMagick dependency and focus on orchestration logic.

### Testing Commands

**Run All Tests**:
```bash
go test ./tests/... -v
```

**Package-Specific Tests**:
```bash
go test ./tests/processing/... -v
go test ./tests/retry/... -v
```

**Coverage Analysis**:
```bash
go test ./tests/... -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Final Architecture

### Package Structure

```
tools/classify-docs/
├── pkg/
│   ├── config/
│   │   ├── cache.go           # Cache configuration with pointer-based Enabled
│   │   ├── config.go          # ClassifyConfig and loading logic
│   │   ├── processing.go      # ProcessingConfig and sub-configs
│   │   └── retry.go           # RetryConfig
│   ├── retry/
│   │   └── retry.go           # Generic retry with exponential backoff
│   ├── cache/
│   │   └── cache.go           # SystemPromptCache persistence
│   ├── document/
│   │   ├── document.go        # Interfaces
│   │   └── pdf.go             # PDF implementation
│   └── processing/
│       ├── progress.go        # Shared ProgressFunc type
│       ├── parallel.go        # ProcessPages with worker pools
│       └── sequential.go      # ProcessWithContext with accumulation
├── tests/                     # Black-box tests mirroring pkg/
├── cmd/
│   ├── test-config/
│   │   └── main.go           # Config verification utility
│   └── test-render/
│       └── main.go           # Image rendering utility
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

**Configuration Types**:
- `ClassifyConfig` - Unified configuration
- `ProcessingConfig` - Processing settings
- `ParallelConfig` - Worker cap
- `SequentialConfig` - Intermediate context exposure
- `RetryConfig` - Retry behavior
- `CacheConfig` - Cache settings with `IsEnabled()` method

**Processing Functions**:
- `ProcessPages[T](...)` - Parallel processing
- `ProcessWithContext[TContext](...)` - Sequential processing

**Retry Function**:
- `Do[T](ctx, cfg, fn)` - Generic retry with exponential backoff
- `MarkNonRetryable(err)` - Mark errors as non-retryable
- `IsRetryable(err)` - Check if error should be retried

**Cache Functions**:
- `New(docs, prompt)` - Create cache entry
- `Load(path)` - Load from disk
- `(*SystemPromptCache).Save(path)` - Save to disk

### Dependencies

**Go Dependencies** (go.mod):
```
require (
    github.com/JaimeStill/go-agents v0.1.0
    github.com/pdfcpu/pdfcpu v0.11.0
)
```

**go-agents Integration**: Imports `pkg/config` for `AgentConfig` and `Duration` types, enabling unified configuration structure.

## Design Validation Results

### Generic Processing Layer

**Assessment**: Processors are successfully agent-agnostic and type-safe.

**Validation**:
- Callbacks delegate domain logic to caller
- Type parameters eliminate casting
- No coupling between processors and specific use cases
- Easy to test with mock processors

### Error Handling

**Assessment**: Comprehensive error coordination across concurrent operations.

**Features**:
- Retry logic handles transient failures
- Context cancellation coordinates fail-fast behavior
- Non-retryable errors prevent wasted attempts
- Error wrapping preserves context with `%w`

### Resource Management

**Assessment**: Clean resource management with proper cleanup.

**Patterns**:
- Worker pools sized based on CPU count and config
- Self-balancing work distribution via channels
- Context cancellation prevents goroutine leaks
- Deferred channel closure ensures completion
- Temporary files cleaned with defer

### Configuration Flexibility

**Assessment**: Minimal config files with sensible defaults.

**Achievements**:
- Can omit all default values
- Pointer-based booleans support default `true`
- Duration fields use human-readable strings
- Easy loading from JSON with merge logic
- Clear separation between config loading and validation

## Performance Characteristics

**Parallel Processing**:
- Worker calculation: `min(min(NumCPU * 2, WorkerCap), len(pages))`
- Default worker cap: 16
- Self-balancing: Fast workers process more pages
- Test measured: 20 pages with 4 workers ~250ms (vs 1000ms sequential)

**Sequential Processing**:
- Linear page-by-page processing
- Context accumulation overhead: minimal
- Optional intermediate context capture

**Retry Behavior**:
- Initial backoff: 1s (default)
- Max backoff: 30s (default)
- Multiplier: 2.0 (exponential growth)
- Test measured: 70ms minimum for 3 attempts with 10ms/20ms/40ms backoffs

**Cache Operations**:
- Load: Milliseconds (JSON parse)
- Save: Milliseconds (JSON serialize + write)
- Automatic directory creation adds negligible overhead

## Lessons Learned

### What Worked Well

1. **Package-Level Generic Functions**: Avoiding method type parameter limitations provided maximum flexibility
2. **Background Result Collection**: Decoupling from worker completion eliminated deadlock
3. **Pre-Cancelled Context Tests**: Made tests instant and deterministic
4. **Pointer-Based Booleans**: Elegant solution for default value merging
5. **Separation of Concerns**: Cache package independent of config package
6. **Modern Go Patterns**: WaitGroup.Go(), range over integer, chained min() improved readability

### What Required Iteration

1. **Deadlock Resolution**: Required architectural change to background collector
2. **Test Reliability**: Moved from timing-dependent to deterministic approach
3. **Boolean Defaults**: Tried conditional merge before discovering pointer solution
4. **Cache Signature**: Removed config dependency to simplify interface

### Recommendations for go-agents-document-context

1. **Generic Processing**: Package-level functions with type parameters are essential
2. **Result Collection**: Background collector pattern handles early worker exit gracefully
3. **Configuration**: Use pointers for booleans that need default `true` behavior
4. **Testing**: Pre-cancelled contexts for deterministic cancellation tests
5. **Channel Safety**: Always use `defer close()` in sender goroutines
6. **Error Coordination**: Context cancellation coordinates fail-fast across workers
7. **Retry Logic**: Simple custom implementation sufficient for most needs

## Current State

### Completed

- ✅ Generic parallel processor with worker pools
- ✅ Generic sequential processor with context accumulation
- ✅ Retry infrastructure with exponential backoff
- ✅ System prompt caching with JSON persistence
- ✅ Unified configuration management
- ✅ Pointer-based boolean defaults
- ✅ Modern Go 1.25.2 patterns (WaitGroup.Go(), etc.)
- ✅ Comprehensive black-box tests (33 tests, all passing)
- ✅ Minimal configuration files (only non-default values)
- ✅ Configuration verification utility
- ✅ Parallel processing deadlock fixed
- ✅ Deterministic context cancellation tests
- ✅ .gitignore for cache directory

### Known Limitations

- No rate limiting (relies on retry backoff)
- No batch size optimization
- No concurrent cache access protection (single-process assumption)
- No cache expiration or validation
- Sequential processing doesn't support cancellation mid-page (only between pages)

### Files Delivered

**Package Files**:
1. `pkg/config/cache.go` - Cache configuration
2. `pkg/config/config.go` - Unified configuration
3. `pkg/config/processing.go` - Processing configurations
4. `pkg/config/retry.go` - Retry configuration
5. `pkg/retry/retry.go` - Retry infrastructure
6. `pkg/cache/cache.go` - Cache persistence
7. `pkg/processing/progress.go` - Shared types
8. `pkg/processing/parallel.go` - Parallel processor
9. `pkg/processing/sequential.go` - Sequential processor

**Test Files**:
10. `tests/config/config_test.go` - Configuration tests
11. `tests/retry/retry_test.go` - Retry tests
12. `tests/cache/cache_test.go` - Cache tests
13. `tests/processing/parallel_test.go` - Parallel processing tests
14. `tests/processing/sequential_test.go` - Sequential processing tests

**Configuration Files**:
15. `config.classify-gpt4o-key.json` - Azure OpenAI with API key
16. `config.classify-gpt4o-entra.json` - Azure OpenAI with Entra ID
17. `config.classify-gemma.json` - Ollama with gemma3:4b

**Utilities**:
18. `cmd/test-config/main.go` - Configuration verification tool

**Documentation**:
19. `.gitignore` - Ignore cache directory and build artifacts
20. `_context/.archive/02-processing-infrastructure.md` - This document

### Next Phase Prerequisites

Phase 4 (System Prompt Generation) depends on:
- ✅ Sequential processing with context accumulation
- ✅ System prompt caching
- ✅ Retry infrastructure for LLM calls
- ✅ Configuration management
- ✅ Document processing primitives

Phase 4 will implement:
- Reference document processing workflow
- System prompt generation from classification guides
- LLM integration for document analysis
- Cache utilization for avoiding reprocessing

Phase 5 (Document Classification) depends on:
- ✅ Parallel processing with worker pools
- ✅ Retry infrastructure for LLM calls
- ✅ Configuration management
- ✅ Generated system prompts (from Phase 4)

Phase 5 will implement:
- Parallel page classification workflow
- Classification result parsing and aggregation
- CLI for document classification
- Output formatting

## Reference Information

### Configuration Loading

**Load Configuration**:
```go
cfg, err := config.LoadClassifyConfig("config.classify-gpt4o-key.json")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}
```

**Access Settings**:
```go
// Agent settings
agentName := cfg.Agent.Name
modelName := cfg.Agent.Transport.Provider.Model.Name

// Processing settings (with defaults)
workerCap := cfg.Processing.Parallel.WorkerCap  // 16 if not specified
maxAttempts := cfg.Processing.Retry.MaxAttempts  // 3 if not specified
cacheEnabled := cfg.Processing.Cache.IsEnabled()  // true if not specified
cachePath := cfg.Processing.Cache.Path  // ".cache/system-prompt.json" if not specified
```

### Parallel Processing

**Basic Usage**:
```go
results, err := processing.ProcessPages(
    ctx,
    cfg.Processing.Parallel,
    pages,
    func(ctx context.Context, page document.Page) (Classification, error) {
        // Process page...
        return classification, nil
    },
    func(completed, total int) {
        log.Printf("Processed %d/%d pages", completed, total)
    },
)
```

**With Retry**:
```go
processFunc := func(ctx context.Context, page document.Page) (Classification, error) {
    return retry.Do(ctx, cfg.Processing.Retry,
        func(ctx context.Context, attempt int) (Classification, error) {
            // Attempt processing...
        },
    )
}

results, err := processing.ProcessPages(ctx, cfg.Processing.Parallel, pages, processFunc, nil)
```

### Sequential Processing

**Basic Usage**:
```go
result, err := processing.ProcessWithContext(
    ctx,
    cfg.Processing.Sequential,
    pages,
    "initial context",
    func(ctx context.Context, page document.Page, prevContext string) (string, error) {
        // Process page and update context...
        return updatedContext, nil
    },
    nil,
)

finalContext := result.Final
```

**With Intermediate Contexts**:
```go
cfg := config.SequentialConfig{ExposeIntermediateContexts: true}

result, err := processing.ProcessWithContext(ctx, cfg, pages, initial, processor, nil)

// Access intermediate states
for i, intermediateContext := range result.Intermediate {
    log.Printf("After page %d: %s", i, intermediateContext)
}
```

### Caching

**Check and Load Cache**:
```go
if cfg.Processing.Cache.IsEnabled() {
    cache, err := cache.Load(cfg.Processing.Cache.Path)
    if err == nil {
        log.Printf("Using cached system prompt from %s", cache.GeneratedAt)
        return cache.SystemPrompt, nil
    }
}
```

**Generate and Save Cache**:
```go
systemPrompt := "Generated prompt..."
refDocs := []string{"doc1.pdf", "doc2.pdf"}

c := cache.New(refDocs, systemPrompt)
if err := c.Save(cfg.Processing.Cache.Path); err != nil {
    log.Printf("Warning: Failed to cache: %v", err)
}
```

### Testing Commands

**Run All Tests**:
```bash
cd tools/classify-docs
go test ./tests/...
```

**Run with Coverage**:
```bash
go test ./tests/... -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

**Verify Configuration**:
```bash
go run ./cmd/test-config/main.go config.classify-gpt4o-key.json
```

### Common Issues and Solutions

**Issue**: Cache shows `Enabled: false` when not specified in config
**Solution**: Use `cfg.Processing.Cache.IsEnabled()` method, not direct field access

**Issue**: Test hangs in parallel processing
**Solution**: Ensure all code paths send to result channel or check for deadlock in collector

**Issue**: Flaky cancellation tests
**Solution**: Use pre-cancelled context: `ctx, cancel := context.WithCancel(ctx); cancel()`

**Issue**: Default values not applying
**Solution**: Verify `Merge()` methods only overwrite non-zero values (or use pointers for booleans)

## Conclusion

Phases 2 and 3 successfully established production-ready processing infrastructure with clean abstractions, comprehensive error handling, and flexible configuration management. The implementation provides generic, reusable processors that work with any result or context type, integrated retry logic for resilience, and system prompt caching for efficiency.

Key achievements:
- **Generic Processing**: Type-safe parallel and sequential processors without runtime casting
- **Deadlock-Free Concurrency**: Background result collector handles early worker termination
- **Flexible Configuration**: Minimal config files with pointer-based defaults
- **Modern Go Patterns**: Leverages Go 1.25.2 features for cleaner, more maintainable code
- **Comprehensive Testing**: 33 black-box tests providing confidence in public API
- **Clean Separation**: Cache package independent of config, processors independent of domain logic

The infrastructure is ready for integration with LLM-based workflows in Phases 4 and 5, with patterns validated through extensive testing and lessons documented for the go-agents-document-context library.
