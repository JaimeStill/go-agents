# Upgrading go-agents-orchestration to go-agents v0.3.0

This document details the changes required to upgrade the `go-agents-orchestration` project from go-agents v0.2.x to v0.3.0.

## Summary of Breaking Changes

go-agents v0.3.0 introduces two major changes:

1. **Package restructuring**: `pkg/types` split into `pkg/protocol`, `pkg/response`, `pkg/model`, and `pkg/request`
2. **Flattened configuration hierarchy**: `Provider` and `Model` are now peer fields with `Client` (not nested under `Client`)

## Files Requiring Code Changes

### 1. examples/darpa-procurement/agents.go

**Current code (lines 6-9):**
```go
import (
    "fmt"

    "github.com/JaimeStill/go-agents/pkg/agent"
    "github.com/JaimeStill/go-agents/pkg/config"
    "github.com/JaimeStill/go-agents/pkg/types"
)
```

**Updated code:**
```go
import (
    "fmt"

    "github.com/JaimeStill/go-agents/pkg/agent"
    "github.com/JaimeStill/go-agents/pkg/config"
    "github.com/JaimeStill/go-agents/pkg/protocol"
)
```

**Current code (line 95):**
```go
if wc.MaxTokens > 0 {
    a.Model().Options[types.Chat]["max_tokens"] = wc.MaxTokens
}
```

**Updated code:**
```go
if wc.MaxTokens > 0 {
    a.Model().Options[protocol.Chat]["max_tokens"] = wc.MaxTokens
}
```

### 2. examples/darpa-procurement/workflow.go

**Current code (lines 4-11):**
```go
import (
    "context"
    "fmt"

    "github.com/JaimeStill/go-agents/pkg/types"
    "github.com/JaimeStill/go-agents-orchestration/pkg/config"
    "github.com/JaimeStill/go-agents-orchestration/pkg/state"
    "github.com/JaimeStill/go-agents-orchestration/pkg/workflows"
)
```

**Updated code:**
```go
import (
    "context"
    "fmt"

    "github.com/JaimeStill/go-agents/pkg/response"
    "github.com/JaimeStill/go-agents-orchestration/pkg/config"
    "github.com/JaimeStill/go-agents-orchestration/pkg/state"
    "github.com/JaimeStill/go-agents-orchestration/pkg/workflows"
)
```

**Current code (line 315):**
```go
processor := func(ctx context.Context, task AnalysisTask) (AnalysisResult, error) {
    var response *types.ChatResponse
    var err error
```

**Updated code:**
```go
processor := func(ctx context.Context, task AnalysisTask) (AnalysisResult, error) {
    var resp *response.ChatResponse
    var err error
```

**Note**: The variable is renamed from `response` to `resp` to avoid collision with the imported package name.

**Current code (line 564-566):**
```go
func routeToExecutive(ctx context.Context, s state.State, executive interface {
    Chat(context.Context, string, ...map[string]any) (*types.ChatResponse, error)
}, title string, cost int, route string) (state.State, error) {
```

**Updated code:**
```go
func routeToExecutive(ctx context.Context, s state.State, executive interface {
    Chat(context.Context, string, ...map[string]any) (*response.ChatResponse, error)
}, title string, cost int, route string) (state.State, error) {
```

## Configuration Files Requiring Updates

All 10 configuration files use the old nested structure and need to be updated to the flattened hierarchy.

### examples/darpa-procurement/config.gemma.json

**Current structure:**
```json
{
  "name": "gemma-agent",
  "client": {
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434",
      "model": {
        "name": "gemma3:4b",
        "capabilities": {
          "chat": {
            "max_tokens": 4096
          }
        }
      }
    }
  }
}
```

**Updated structure:**
```json
{
  "name": "gemma-agent",
  "client": {},
  "provider": {
    "name": "ollama",
    "base_url": "http://localhost:11434"
  },
  "model": {
    "name": "gemma3:4b",
    "capabilities": {
      "chat": {
        "max_tokens": 4096
      }
    }
  }
}
```

### examples/phase-01-hubs/config.gemma.json

**Updated structure:**
```json
{
  "name": "gemma-agent",
  "client": {},
  "provider": {
    "name": "ollama",
    "base_url": "http://localhost:11434"
  },
  "model": {
    "name": "gemma3:4b",
    "capabilities": {
      "chat": {
        "max_tokens": 150
      },
      "vision": {
        "max_tokens": 150
      }
    }
  }
}
```

### examples/phase-01-hubs/config.llama.json

Apply the same flattening pattern - move `provider` and `model` to be peers with `client`.

### examples/phase-02-03-state-graphs/config.llama.json

Apply the same flattening pattern.

### examples/phase-04-sequential-chains/config.llama.json

Apply the same flattening pattern.

### examples/phase-05-parallel-execution/config.llama.json

Apply the same flattening pattern.

### examples/phase-06-checkpointing/config.llama.json

Apply the same flattening pattern.

### examples/phase-07-conditional-routing/config.llama.json

Apply the same flattening pattern.

### examples/phase-07-conditional-routing/config.gemma.json

Apply the same flattening pattern.

## No Changes Required

The following packages use go-agents interfaces that remain compatible:

- **pkg/hub/hub.go** - Uses `agent.Agent` interface which is unchanged
- **pkg/config/** - Uses `config.AgentConfig` which handles the new structure automatically via `Merge()`
- **pkg/state/** - Does not directly use go-agents types
- **pkg/workflows/** - Does not directly use go-agents types
- **pkg/messaging/** - Does not use go-agents types

## Verification Steps

After making the changes:

1. Update go.mod dependency:
   ```bash
   go get github.com/JaimeStill/go-agents@v0.3.0
   go mod tidy
   ```

2. Run build to verify compilation:
   ```bash
   go build ./...
   ```

3. Run tests:
   ```bash
   go test ./tests/...
   ```

4. Test with Ollama example:
   ```bash
   cd examples/darpa-procurement
   go run . -config config.gemma.json
   ```

## Type Reference Quick Reference

| Old Type (v0.2.x) | New Type (v0.3.0) | Package |
|-------------------|-------------------|---------|
| `types.Chat` | `protocol.Chat` | `pkg/protocol` |
| `types.Vision` | `protocol.Vision` | `pkg/protocol` |
| `types.Tools` | `protocol.Tools` | `pkg/protocol` |
| `types.Embeddings` | `protocol.Embeddings` | `pkg/protocol` |
| `*types.ChatResponse` | `*response.ChatResponse` | `pkg/response` |
| `*types.VisionResponse` | `*response.ChatResponse` | `pkg/response` |
| `*types.ToolsResponse` | `*response.ToolsResponse` | `pkg/response` |
| `*types.EmbeddingsResponse` | `*response.EmbeddingsResponse` | `pkg/response` |
| `*types.StreamingChunk` | `*response.StreamingChunk` | `pkg/response` |
| `types.Message` | `protocol.Message` | `pkg/protocol` |
