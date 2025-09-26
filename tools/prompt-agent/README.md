# prompt-agent

A CLI tool for testing the Agent package primitives by sending prompts to configured LLM providers.

## Usage

```bash
go run tools/prompt-agent/main.go -config <config-file> -prompt <prompt> [options]
```

### Required Flags

- `-config`: Path to JSON configuration file (default: "config.ollama.json")
- `-prompt`: The prompt text to send to the agent

### Optional Flags

- `-system-prompt`: Override the system prompt (takes precedence over config file)
- `-token`: Authentication token (API key or bearer token, depending on auth_type)
- `-stream`: Use ChatStream instead of Chat method

## Examples

### Basic Usage

Send a simple prompt using default Ollama configuration:

```bash
go run tools/prompt-agent/main.go -prompt "Hello, how are you?"
```

### With System Prompt

Override the system prompt for specific behavior:

```bash
go run tools/prompt-agent/main.go \
  -system-prompt "You are a helpful math assistant. Answer only with the numerical result." \
  -prompt "What is 2 + 2?"
```

### With Custom Configuration File

Use a custom configuration file:

```bash
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "Tell me about yourself"
```

### Configuration with System Prompt Override

Load configuration from file but override the system prompt:

```bash
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -system-prompt "You are a pirate. Speak like one." \
  -prompt "Tell me about the weather"
```

### Streaming Response

Use streaming for real-time response:

```bash
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "Tell me a story" \
  -stream
```

### Azure with API Key

Use Azure with API key authentication:

```bash
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure.json \
  -token "your-api-key-here" \
  -prompt "Describe the benefits of the Go programming language"
```

### Azure with Bearer Token Authentication

Use Azure Entra ID for authentication:

```bash
# Get Azure bearer token
AZURE_TOKEN=$(. scripts/azure/utilities/get-foundry-token.sh)

# Use token with Azure configuration
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure-entra.json \
  -token $AZURE_TOKEN \
  -prompt "Describe the benefits of the Go programming language" \
  -stream
```

## Configuration File Format

The configuration file follows the structure defined in `pkg/config/agent.go`:

### Ollama Configuration

```json
{
  "name": "ollama-agent",
  "system_prompt": "You are a helpful assistant",
  "client": {
    "provider": "ollama",
    "format": "openai-standard",
    "endpoint": "http://localhost:11434",
    "model": "llama3.2:3b",
    "timeout": 60000000000,
    "max_retries": 3,
    "retry_backoff_base": 1000000000,
    "connection_pool_size": 10,
    "connection_timeout": 90000000000,
    "options": {
      "max_tokens": 4096,
      "temperature": 0.7,
      "top_p": 0.95
    }
  }
}
```

### Azure with API Key Authentication

```json
{
  "name": "azure-key-agent",
  "client": {
    "provider": "azure",
    "format": "openai-reasoning",
    "endpoint": "https://go-agents-platform.openai.azure.com",
    "model": "o3-mini",
    "options": {
      "deployment": "o3-mini",
      "api_version": "2025-01-01-preview",
      "auth_type": "api_key",
      "max_completion_tokens": 4096
    }
  }
}
```

### Azure with Bearer Token Authentication

```json
{
  "name": "azure-token-agent",
  "client": {
    "provider": "azure",
    "format": "openai-reasoning",
    "endpoint": "https://go-agents-platform.openai.azure.com",
    "model": "o3-mini",
    "options": {
      "deployment": "o3-mini",
      "api_version": "2025-01-01-preview",
      "auth_type": "bearer",
      "max_completion_tokens": 4096
    }
  }
}
```

**Notes**:
- Timeout values are in nanoseconds (Go's time.Duration format)
- Format-specific parameters are now in the `options` map:
  - **openai-standard**: Uses `max_tokens`, `temperature`, `top_p`
  - **openai-reasoning**: Uses `max_completion_tokens` only (ignores temperature/top_p)
- Azure requires custom subdomain endpoints for bearer token authentication
- The `auth_type` option supports `"api_key"` or `"bearer"` for Azure provider
- Authentication credentials are provided via the `-token` command line flag

## Output

The tool outputs the raw `CompletionResponse` struct, which includes:
- Model name
- Response text
- Completion status
- Context tokens (for providers that support it)
- Performance metrics (total time, load time, token count, etc.)

## Authentication Methods

### API Key Authentication
- Set `options.auth_type` to `"api_key"` in configuration file
- Provide API key via `-token` command line flag
- Works with both regional and custom subdomain endpoints

### Bearer Token Authentication (Azure Entra ID)
- Set `options.auth_type` to `"bearer"` in configuration file
- Provide bearer token via `-token` command line flag
- Requires custom subdomain endpoints (e.g., `https://your-domain.openai.azure.com`)
- Token can be obtained using `scripts/azure/utilities/get-foundry-token.sh`
- Requires appropriate Azure role assignments (handled by infrastructure scripts)

## Default Configuration

When no configuration file is specified, the tool uses the default configuration from `pkg/agent.DefaultConfig()`, which connects to Ollama on `http://localhost:11434` using the `llama3.2:3b` model.
