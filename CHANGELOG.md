# Changelog

## [v0.1.3] - 2025-10-23

**Changed**:
- Capability format naming: renamed from vendor-centric to specification-based naming
  - `openai-chat` → `chat` (standard OpenAI-compatible chat completions)
  - `openai-vision` → `vision` (standard OpenAI-compatible vision API)
  - `openai-tools` → `tools` (standard OpenAI-compatible function calling)
  - `openai-embeddings` → `embeddings` (standard OpenAI-compatible embeddings)
  - `openai-reasoning` → `o-chat` (OpenAI o-series reasoning models)

**Added**:
- `o-vision` capability format for OpenAI o-series vision reasoning models
  - Supports `max_completion_tokens`, `reasoning_effort`, `images`, `detail` parameters
  - Uses o-series parameter restrictions (no temperature/top_p support)

## [v0.1.2] - 2025-10-10

**Added**:
- `pkg/mock` package providing mock implementations for testing
- `MockAgent`, `MockClient`, `MockProvider`, `MockModel`, `MockCapability` types
- Helper constructors: `NewSimpleChatAgent`, `NewStreamingChatAgent`, `NewToolsAgent`, `NewEmbeddingsAgent`, `NewMultiProtocolAgent`, `NewFailingAgent`

## [v0.1.1] - 2025-10-10

**Added**:
- `ID() string` method to Agent interface returning unique UUIDv7 identifier

## [v0.1.0] - 2025-10-09

Initial pre-release.

**Protocols**:
- Chat
- Vision
- Tools
- Embeddings

**Capability Formats**:
- openai-chat
- openai-vision
- openai-tools
- openai-embeddings
- openai-reasoning

**Providers**:
- Ollama
- Azure AI Foundry
