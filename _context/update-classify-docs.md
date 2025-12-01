# Upgrading classify-docs to go-agents v0.3.0

This document details the changes required to upgrade the `tools/classify-docs` project from go-agents v0.2.x to v0.3.0.

## Summary of Breaking Changes

go-agents v0.3.0 introduces a flattened configuration hierarchy:
- `AgentConfig.Provider` is now a peer field with `Client` (not nested under `Client`)
- `AgentConfig.Model` is now a peer field with `Client` (not nested under `Client.Provider`)
- `ClientConfig` contains only HTTP settings (timeout, retry, connection settings)
- `ProviderConfig` contains only provider connection settings (name, base_url, options)

## Files Requiring Changes

### 1. main.go

**Current code (line 61-62):**
```go
if *token != "" {
    cfg.Agent.Client.Provider.Options["token"] = *token
}
```

**Updated code:**
```go
if *token != "" {
    if cfg.Agent.Provider.Options == nil {
        cfg.Agent.Provider.Options = make(map[string]any)
    }
    cfg.Agent.Provider.Options["token"] = *token
}
```

**Current code (line 107-108):**
```go
if *token != "" {
    cfg.Agent.Client.Provider.Options["token"] = *token
}
```

**Updated code:**
```go
if *token != "" {
    if cfg.Agent.Provider.Options == nil {
        cfg.Agent.Provider.Options = make(map[string]any)
    }
    cfg.Agent.Provider.Options["token"] = *token
}
```

### 2. config.classify-gpt4o-key.json

**Current structure:**
```json
{
  "agent": {
    "name": "classify-agent-gpt4o",
    "client": {
      "provider": {
        "name": "azure",
        "base_url": "https://go-agents-platform.openai.azure.com/openai",
        "model": {
          "name": "gpt-4o",
          "capabilities": {
            "vision": {
              "max_tokens": 4096,
              "temperature": 0.1,
              "vision_options": {
                "detail": "high"
              }
            }
          }
        },
        "options": {
          "deployment": "gpt-4o",
          "api_version": "2025-01-01-preview",
          "auth_type": "api_key"
        }
      }
    }
  }
}
```

**Updated structure:**
```json
{
  "agent": {
    "name": "classify-agent-gpt4o",
    "client": {},
    "provider": {
      "name": "azure",
      "base_url": "https://go-agents-platform.openai.azure.com/openai",
      "options": {
        "deployment": "gpt-4o",
        "api_version": "2025-01-01-preview",
        "auth_type": "api_key"
      }
    },
    "model": {
      "name": "gpt-4o",
      "capabilities": {
        "vision": {
          "max_tokens": 4096,
          "temperature": 0.1,
          "vision_options": {
            "detail": "high"
          }
        }
      }
    }
  }
}
```

### 3. config.classify-gpt4o-entra.json

Same structural change as above, with `auth_type: "bearer"`.

### 4. config.classify-o4-mini.json

**Updated structure:**
```json
{
  "agent": {
    "name": "classify-agent-o4mini",
    "client": {},
    "provider": {
      "name": "azure",
      "base_url": "https://go-agents-platform.openai.azure.com/openai",
      "options": {
        "deployment": "o4-mini",
        "api_version": "2025-01-01-preview",
        "auth_type": "api_key"
      }
    },
    "model": {
      "name": "o4-mini",
      "capabilities": {
        "vision": {
          "reasoning_effort": "high",
          "vision_options": {
            "detail": "high"
          }
        }
      }
    }
  }
}
```

### 5. config.classify-gemma.json

**Updated structure:**
```json
{
  "agent": {
    "name": "classify-agent-gemma",
    "client": {},
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434"
    },
    "model": {
      "name": "gemma3:4b",
      "capabilities": {
        "vision": {
          "max_tokens": 4096,
          "temperature": 0.1,
          "vision_options": {
            "detail": "auto"
          }
        }
      }
    }
  }
}
```

### 6. config.classify-gpt-5-mini.json

**Updated structure:**
```json
{
  "agent": {
    "name": "classify-agent-gpt5mini",
    "client": {},
    "provider": {
      "name": "azure",
      "base_url": "https://go-agents-platform.openai.azure.com/openai",
      "options": {
        "deployment": "gpt-5-mini",
        "api_version": "2025-01-01-preview",
        "auth_type": "api_key"
      }
    },
    "model": {
      "name": "gpt-5-mini",
      "capabilities": {
        "vision": {
          "max_tokens": 4096,
          "temperature": 0.1,
          "vision_options": {
            "detail": "high"
          }
        }
      }
    }
  }
}
```

## No Changes Required

The following files use go-agents types that remain compatible:

- **pkg/prompt/prompt.go** - Uses `agent.Agent` interface and `response.Content()` method which are unchanged
- **pkg/classify/classify.go** - Uses `agent.Agent.Vision()` method which is unchanged
- **pkg/config/config.go** - Uses `config.AgentConfig` which handles the new structure automatically via `Merge()`

## Verification Steps

After making the changes:

1. Run `go build ./...` to verify compilation
2. Test with Ollama: `go run . generate-prompt -config config.classify-gemma.json -references _context`
3. Test with Azure: `go run . classify -config config.classify-gpt4o-key.json -token $AZURE_API_KEY -input _context/marked-documents`
