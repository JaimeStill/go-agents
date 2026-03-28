// Package response defines the provider-neutral types returned by LLM protocol requests.
//
// The core type is [Response], which represents the result of Chat, Vision, and Tools
// protocol requests using typed content blocks. [TextBlock] carries text content,
// [ToolUseBlock] carries tool call requests, and a single response can contain both.
// Convenience methods [Response.Text] and [Response.ToolCalls] extract content by type.
//
// [StreamingResponse] represents partial responses delivered as streaming chunks,
// using the same content block model. The Error field signals transport errors.
//
// [EmbeddingsResponse] is a separate type for the Embeddings protocol, which returns
// vector data rather than conversational content.
//
// [TokenUsage] normalizes token consumption across providers with InputTokens,
// OutputTokens, and TotalTokens fields.
package response
