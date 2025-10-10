# Changelog

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
