package response

// StreamingResponse holds the accumulated state of a streaming model response.
type StreamingResponse struct {
	Content    []ContentBlock
	StopReason string
	Usage      *TokenUsage
	Error      error
}

// Text returns the concatenated text from all TextBlock content blocks.
func (r *StreamingResponse) Text() string {
	var text string
	for _, block := range r.Content {
		if tb, ok := block.(TextBlock); ok {
			text += tb.Text
		}
	}
	return text
}
