# Classify-Docs Component Extraction

Complete implementation guide for extracting classify-docs infrastructure into standalone libraries and production project.

## Overview

This guide covers the complete extraction workflow:

1. **go-agents-document-context library** - Standalone document processing library (zero go-agents dependency)
2. **classify-docs production project** - Standalone classification tool consuming go-agents and document-context

The extraction separates reusable document infrastructure from tool-specific classification logic, enabling:
- Document processing library reuse across projects
- Cleaner dependency boundaries
- Independent versioning and evolution
- Production-ready classify-docs tool

---

## Prerequisites

- Go 1.25.2+
- Git
- GitHub account (for repository creation)
- Existing go-agents v0.2.0+ installation
- Completed classify-docs prototype in `~/code/go-agents/tools/classify-docs/`

---

## Part 1: Create go-agents-document-context Library

### Goal

Extract document processing infrastructure into standalone library with zero go-agents dependency.

### Step 1.1: Initialize Repository

```bash
# Create repository directory
cd ~/code
mkdir go-agents-document-context
cd go-agents-document-context

# Initialize git
git init
git branch -M main

# Create directory structure
mkdir -p pkg/document
mkdir -p pkg/encoding
mkdir -p pkg/processing
mkdir -p tests/document
mkdir -p tests/encoding
mkdir -p tests/processing
mkdir -p examples

# Initialize Go module
go mod init github.com/JaimeStill/go-agents-document-context
```

### Step 1.2: Extract Document Package

**Copy source files:**

```bash
# From go-agents/tools/classify-docs/
cp ~/code/go-agents/tools/classify-docs/pkg/document/*.go ./pkg/document/
```

**Update imports in pkg/document/*.go:**

```go
// Remove any classify-docs internal imports
// Keep only standard library and pdfcpu imports

import (
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)
```

**Add pdfcpu dependency:**

```bash
go get github.com/pdfcpu/pdfcpu/pkg/api@latest
go get github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model@latest
```

**Copy tests:**

```bash
cp ~/code/go-agents/tools/classify-docs/tests/document/*.go ./tests/document/
```

**Update test package declaration:**

```go
package document_test

import (
	"testing"

	"github.com/JaimeStill/go-agents-document-context/pkg/document"
)
```

**Verify compilation:**

```bash
go build ./pkg/document/
go test ./tests/document/... -v
```

### Step 1.3: Extract Encoding Package

**Copy source files:**

```bash
cp ~/code/go-agents/tools/classify-docs/pkg/encoding/*.go ./pkg/encoding/
```

**Update imports (should be standard library only):**

```go
import (
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)
```

**Copy tests:**

```bash
cp ~/code/go-agents/tools/classify-docs/tests/encoding/*.go ./tests/encoding/
```

**Update test imports:**

```go
package encoding_test

import (
	"testing"

	"github.com/JaimeStill/go-agents-document-context/pkg/encoding"
)
```

**Verify compilation:**

```bash
go build ./pkg/encoding/
go test ./tests/encoding/... -v
```

### Step 1.4: Extract Processing Package

**Copy source files:**

```bash
cp ~/code/go-agents/tools/classify-docs/pkg/processing/*.go ./pkg/processing/
```

**Update imports (keep generic, no go-agents dependency):**

```go
import (
	"context"
	"fmt"
	"sync"
)
```

**Important:** The processing package should remain completely generic. If any go-agents-specific logic exists, remove it.

**Copy tests:**

```bash
cp ~/code/go-agents/tools/classify-docs/tests/processing/*.go ./tests/processing/
```

**Update test imports:**

```go
package processing_test

import (
	"context"
	"testing"

	"github.com/JaimeStill/go-agents-document-context/pkg/processing"
)
```

**Verify compilation:**

```bash
go build ./pkg/processing/
go test ./tests/processing/... -v
```

### Step 1.5: Create Examples

**Example 1: PDF Reading** (`examples/pdf-reading/main.go`)

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/JaimeStill/go-agents-document-context/pkg/document"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <pdf-file>")
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to open PDF: %v", err)
	}
	defer file.Close()

	doc, err := document.NewPDFDocument(file)
	if err != nil {
		log.Fatalf("Failed to create PDF document: %v", err)
	}

	fmt.Printf("PDF Pages: %d\n", doc.PageCount())

	// Read first page
	page, err := doc.Page(0)
	if err != nil {
		log.Fatalf("Failed to read page: %v", err)
	}

	fmt.Printf("Page %d image size: %dx%d\n",
		page.Number(),
		page.Image().Bounds().Dx(),
		page.Image().Bounds().Dy())
}
```

**Example 2: Data URI Encoding** (`examples/data-uri/main.go`)

```go
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/JaimeStill/go-agents-document-context/pkg/encoding"
)

func main() {
	// Create simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Encode as data URI
	dataURI, err := encoding.EncodeImageDataURI(img, "png")
	if err != nil {
		log.Fatalf("Failed to encode: %v", err)
	}

	fmt.Printf("Data URI (first 100 chars): %s...\n", dataURI[:100])
	fmt.Printf("Full length: %d characters\n", len(dataURI))
}
```

**Example 3: Sequential Processing** (`examples/sequential-processing/main.go`)

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/JaimeStill/go-agents-document-context/pkg/processing"
)

type AccumulatedContext struct {
	Values []int
	Sum    int
}

func main() {
	ctx := context.Background()

	processor := processing.NewContextProcessor(
		[]int{1, 2, 3, 4, 5},
		&AccumulatedContext{Values: make([]int, 0)},
		func(ctx context.Context, item int, accum *AccumulatedContext) error {
			accum.Values = append(accum.Values, item)
			accum.Sum += item
			return nil
		},
	)

	result, err := processor.ProcessWithContext(ctx, nil)
	if err != nil {
		log.Fatalf("Processing failed: %v", err)
	}

	fmt.Printf("Processed values: %v\n", result.Context.Values)
	fmt.Printf("Sum: %d\n", result.Context.Sum)
	fmt.Printf("Total items: %d\n", len(result.Results))
}
```

### Step 1.6: Create README.md

```markdown
# go-agents-document-context

Document processing infrastructure for AI agent workflows.

## Features

- **Document Abstraction**: Generic document and page interfaces
- **PDF Support**: PDF reading via pdfcpu with image extraction
- **Image Encoding**: Data URI encoding for vision API compatibility
- **Sequential Processing**: Generic context accumulation framework

## Installation

\`\`\`bash
go get github.com/JaimeStill/go-agents-document-context
\`\`\`

## Packages

### pkg/document

Document and page abstractions with PDF implementation.

\`\`\`go
import "github.com/JaimeStill/go-agents-document-context/pkg/document"

file, _ := os.Open("document.pdf")
doc, _ := document.NewPDFDocument(file)

page, _ := doc.Page(0)
img := page.Image()
\`\`\`

### pkg/encoding

Image to data URI encoding for vision APIs.

\`\`\`go
import "github.com/JaimeStill/go-agents-document-context/pkg/encoding"

dataURI, _ := encoding.EncodeImageDataURI(img, "png")
// Returns: data:image/png;base64,iVBORw0KGgo...
\`\`\`

### pkg/processing

Generic sequential processing with context accumulation.

\`\`\`go
import "github.com/JaimeStill/go-agents-document-context/pkg/processing"

type Context struct { Sum int }

processor := processing.NewContextProcessor(
    items,
    &Context{},
    func(ctx context.Context, item int, accum *Context) error {
        accum.Sum += item
        return nil
    },
)

result, _ := processor.ProcessWithContext(ctx, nil)
\`\`\`

## Dependencies

- **pdfcpu**: PDF processing
- **Standard library only** for encoding and processing

## Testing

\`\`\`bash
go test ./tests/... -v
\`\`\`

## Examples

See `examples/` directory for complete usage examples.

## License

MIT
```

### Step 1.7: Add LICENSE

```bash
# Create MIT license (adjust year and name as needed)
cat > LICENSE << 'EOF'
MIT License

Copyright (c) 2025 Jaime Still

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF
```

### Step 1.8: Final Verification

```bash
# Clean and verify all dependencies
go mod tidy

# Build all packages
go build ./...

# Run all tests
go test ./tests/... -v

# Run examples
go run examples/data-uri/main.go
go run examples/sequential-processing/main.go

# Verify zero go-agents dependency
go list -m all | grep go-agents
# Should show ONLY go-agents-document-context, not go-agents
```

### Step 1.9: Create GitHub Repository and Push

```bash
# Create .gitignore
cat > .gitignore << 'EOF'
# Binaries
*.exe
*.dll
*.so
*.dylib

# Test binaries
*.test

# Coverage
*.out

# IDE
.idea/
.vscode/
*.swp
*.swo
*~
EOF

# Initial commit
git add .
git commit -m "Initial commit: Document context library v0.1.0"

# Create GitHub repo (via gh CLI or web interface)
gh repo create go-agents-document-context --public --source=. --remote=origin --push

# Tag v0.1.0
git tag v0.1.0
git push origin v0.1.0
```

**Repository URL:** `https://github.com/JaimeStill/go-agents-document-context`

---

## Part 2: Create Standalone classify-docs Project

### Goal

Create production-ready classify-docs tool consuming go-agents and document-context libraries.

### Step 2.1: Initialize Repository

```bash
cd ~/code
mkdir classify-docs
cd classify-docs

# Initialize git
git init
git branch -M main

# Create directory structure
mkdir -p cmd/classify-docs
mkdir -p pkg/classifier
mkdir -p pkg/prompts
mkdir -p pkg/cache
mkdir -p configs
mkdir -p tests/classifier
mkdir -p tests/prompts
mkdir -p tests/cache

# Initialize Go module
go mod init github.com/JaimeStill/classify-docs
```

### Step 2.2: Add Dependencies

```bash
# Add go-agents dependency
go get github.com/JaimeStill/go-agents@v0.2.0

# Add document-context dependency
go get github.com/JaimeStill/go-agents-document-context@v0.1.0
```

### Step 2.3: Migrate Classifier Package

**Copy source files:**

```bash
cp ~/code/go-agents/tools/classify-docs/pkg/classify/*.go ./pkg/classifier/
```

**Update package name:**

```go
package classifier  // was: package classify
```

**Update imports:**

```go
import (
	"context"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents-document-context/pkg/document"
	"github.com/JaimeStill/go-agents-document-context/pkg/processing"
)
```

**Copy tests:**

```bash
cp ~/code/go-agents/tools/classify-docs/tests/classify/*.go ./tests/classifier/
```

**Update test package:**

```go
package classifier_test

import (
	"testing"

	"github.com/JaimeStill/classify-docs/pkg/classifier"
)
```

**Verify:**

```bash
go build ./pkg/classifier/
go test ./tests/classifier/... -v
```

### Step 2.4: Migrate Prompts Package

**Copy source files:**

```bash
cp ~/code/go-agents/tools/classify-docs/pkg/prompt/*.go ./pkg/prompts/
```

**Update package name:**

```go
package prompts  // was: package prompt
```

**Update imports (should be minimal):**

```go
import (
	"fmt"
	"strings"
)
```

**Copy tests:**

```bash
cp ~/code/go-agents/tools/classify-docs/tests/prompt/*.go ./tests/prompts/
```

**Update test package:**

```go
package prompts_test

import (
	"testing"

	"github.com/JaimeStill/classify-docs/pkg/prompts"
)
```

**Verify:**

```bash
go build ./pkg/prompts/
go test ./tests/prompts/... -v
```

### Step 2.5: Migrate Cache Package

**Copy source files:**

```bash
cp ~/code/go-agents/tools/classify-docs/pkg/cache/*.go ./pkg/cache/
```

**Update imports:**

```go
import (
	"context"

	"github.com/JaimeStill/go-agents/pkg/agent"
)
```

**Copy tests:**

```bash
cp ~/code/go-agents/tools/classify-docs/tests/cache/*.go ./tests/cache/
```

**Update test package:**

```go
package cache_test

import (
	"testing"

	"github.com/JaimeStill/classify-docs/pkg/cache"
)
```

**Verify:**

```bash
go build ./pkg/cache/
go test ./tests/cache/... -v
```

### Step 2.6: Create Main CLI

**File:** `cmd/classify-docs/main.go`

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/JaimeStill/classify-docs/pkg/classifier"
	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents-document-context/pkg/document"
)

func main() {
	configFile := flag.String("config", "", "Path to agent configuration file")
	pdfFile := flag.String("pdf", "", "Path to PDF file to classify")
	flag.Parse()

	if *configFile == "" || *pdfFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Load agent configuration
	cfg, err := config.LoadAgentConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create agent
	a, err := agent.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Open PDF
	file, err := os.Open(*pdfFile)
	if err != nil {
		log.Fatalf("Failed to open PDF: %v", err)
	}
	defer file.Close()

	doc, err := document.NewPDFDocument(file)
	if err != nil {
		log.Fatalf("Failed to create PDF document: %v", err)
	}

	// Create classifier
	c := classifier.New(a, doc)

	// Classify document
	ctx := context.Background()
	result, err := c.Classify(ctx)
	if err != nil {
		log.Fatalf("Classification failed: %v", err)
	}

	fmt.Printf("Classification: %s (confidence: %.2f)\n",
		result.Category,
		result.Confidence)
}
```

### Step 2.7: Migrate Configuration Files

```bash
# Copy configuration examples
cp ~/code/go-agents/tools/classify-docs/config*.json ./configs/

# Update any paths or references as needed
```

### Step 2.8: Create README.md

```markdown
# classify-docs

AI-powered document classification tool using vision LLMs.

## Features

- Multi-page PDF support
- Vision-based classification (no OCR required)
- System prompt caching for efficiency
- Sequential page processing with context accumulation
- Support for Ollama and Azure OpenAI

## Installation

\`\`\`bash
go install github.com/JaimeStill/classify-docs/cmd/classify-docs@latest
\`\`\`

## Usage

\`\`\`bash
classify-docs -config config.json -pdf document.pdf
\`\`\`

## Configuration

See `configs/` directory for examples:

- `config.ollama.json` - Local Ollama setup
- `config.azure.json` - Azure OpenAI setup

## Development

\`\`\`bash
# Run tests
go test ./...

# Build
go build ./cmd/classify-docs

# Run locally
./classify-docs -config configs/config.ollama.json -pdf test.pdf
\`\`\`

## Dependencies

- [go-agents](https://github.com/JaimeStill/go-agents) - Agent framework
- [go-agents-document-context](https://github.com/JaimeStill/go-agents-document-context) - Document processing

## License

MIT
```

### Step 2.9: Final Verification

```bash
# Clean dependencies
go mod tidy

# Build everything
go build ./...

# Run all tests
go test ./... -v

# Build CLI
go build -o classify-docs ./cmd/classify-docs/

# Test CLI
./classify-docs -config configs/config.ollama.json -pdf test.pdf
```

### Step 2.10: Create GitHub Repository and Push

```bash
# Create .gitignore
cat > .gitignore << 'EOF'
# Binaries
classify-docs
*.exe
*.dll
*.so
*.dylib

# Test binaries
*.test

# Coverage
*.out

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# Sensitive configs
config.*.local.json
EOF

# Initial commit
git add .
git commit -m "Initial commit: Classify-docs production tool v1.0.0"

# Create GitHub repo
gh repo create classify-docs --public --source=. --remote=origin --push

# Tag v1.0.0
git tag v1.0.0
git push origin v1.0.0
```

**Repository URL:** `https://github.com/JaimeStill/classify-docs`

---

## Part 3: Cleanup Original Prototype

### Step 3.1: Archive Prototype Documentation

```bash
cd ~/code/go-agents/tools/classify-docs

# Move context documents to go-agents archive
mv _context/*.md ~/code/go-agents/_context/.archive/

# Rename with prefix for clarity
cd ~/code/go-agents/_context/.archive/
mv classify-docs-phase1.md 13-classify-docs-phase1.md
mv classify-docs-phase2.md 14-classify-docs-phase2.md
# ... etc for any other classify-docs context docs
```

### Step 3.2: Remove Prototype (Optional)

```bash
# After verifying standalone classify-docs works
cd ~/code/go-agents
rm -rf tools/classify-docs

# Or keep as reference:
mv tools/classify-docs tools/.archive/classify-docs-prototype
```

---

## Part 4: Verification and Testing

### Step 4.1: Integration Test

**Test document-context library independently:**

```bash
cd ~/code/go-agents-document-context
go test ./... -v -cover
```

**Test classify-docs with real documents:**

```bash
cd ~/code/classify-docs

# Test with Ollama
./classify-docs -config configs/config.ollama.json -pdf test-document.pdf

# Test with Azure
export AZURE_API_KEY=$(az account get-access-token --resource https://cognitiveservices.azure.com --query accessToken -o tsv)
./classify-docs -config configs/config.azure.json -pdf test-document.pdf
```

### Step 4.2: Verify Zero Circular Dependencies

```bash
# In document-context: should NOT depend on go-agents
cd ~/code/go-agents-document-context
go list -m all | grep go-agents
# Should only show: github.com/JaimeStill/go-agents-document-context

# In classify-docs: should depend on both libraries
cd ~/code/classify-docs
go list -m all | grep go-agents
# Should show:
# github.com/JaimeStill/go-agents v0.2.0
# github.com/JaimeStill/go-agents-document-context v0.1.0
```

### Step 4.3: Performance Baseline

```bash
cd ~/code/classify-docs

# Time classification
time ./classify-docs -config configs/config.ollama.json -pdf large-document.pdf

# Note baseline metrics:
# - Pages processed per second
# - Total classification time
# - Memory usage
```

---

## Success Criteria

### Document Context Library (v0.1.0)
- [x] Zero go-agents dependency verified
- [x] All tests passing (document, encoding, processing)
- [x] Examples working
- [x] README complete
- [x] GitHub repository created
- [x] v0.1.0 tagged and released

### Classify-Docs Project (v1.0.0)
- [x] Depends on go-agents v0.2.0+
- [x] Depends on document-context v0.1.0
- [x] All tests passing (classifier, prompts, cache)
- [x] CLI working with both Ollama and Azure
- [x] README complete
- [x] GitHub repository created
- [x] v1.0.0 tagged and released

### Cleanup
- [x] Prototype archived or removed
- [x] Documentation migrated
- [x] No broken references

---

## Next Steps

1. **Publish Libraries:**
   - Announce go-agents-document-context v0.1.0
   - Announce classify-docs v1.0.0

2. **Performance Optimization:**
   - Baseline testing with gpt-4o-mini
   - Identify failure patterns
   - Tune classification strategies
   - Optimize confidence thresholds

3. **Future Enhancements:**
   - Add Word document support to document-context
   - Add image document support
   - Parallel page processing
   - Integration with go-agents-orchestration

---

## Troubleshooting

### Import Errors After Extraction

**Problem:** `package X is not in GOROOT`

**Solution:** Run `go mod tidy` in both repositories

### Test Failures After Migration

**Problem:** Tests fail with import errors

**Solution:** Update test package declarations to use new import paths

### Circular Dependency

**Problem:** document-context accidentally depends on go-agents

**Solution:** Review imports, remove go-agents references, use only standard library and pdfcpu

### CLI Not Finding Packages

**Problem:** `classify-docs` binary doesn't run

**Solution:** Ensure `go mod tidy` was run and all dependencies are available
