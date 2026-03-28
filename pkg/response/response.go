package response

// Response represents a complete model response containing content blocks, stop reason, and token usage.
type Response struct {
	Role       string
	Content    []ContentBlock
	StopReason string
	Usage      *TokenUsage
}

// Text returns the concatenated text from all TextBlock content blocks.
func (r *Response) Text() string {
	var text string
	for _, block := range r.Content {
		if tb, ok := block.(TextBlock); ok {
			text += tb.Text
		}
	}
	return text
}

// ToolCalls returns all ToolUseBlock content blocks from the response.
func (r *Response) ToolCalls() []ToolUseBlock {
	var calls []ToolUseBlock
	for _, block := range r.Content {
		if tb, ok := block.(ToolUseBlock); ok {
			calls = append(calls, tb)
		}
	}
	return calls
}
