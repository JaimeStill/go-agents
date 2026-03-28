package response

// TokenUsage tracks token consumption for a request/response cycle.
// Provides counts for prompt tokens, completion tokens, and total tokens used.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
