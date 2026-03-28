package response_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/response"
)

func TestResponse_Text_SingleBlock(t *testing.T) {
	resp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Hello, world!"},
		},
	}

	if resp.Text() != "Hello, world!" {
		t.Errorf("got text %q, want %q", resp.Text(), "Hello, world!")
	}
}

func TestResponse_Text_MultipleBlocks(t *testing.T) {
	resp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Hello, "},
			response.TextBlock{Text: "world!"},
		},
	}

	if resp.Text() != "Hello, world!" {
		t.Errorf("got text %q, want %q", resp.Text(), "Hello, world!")
	}
}

func TestResponse_Text_EmptyContent(t *testing.T) {
	resp := &response.Response{
		Role:    "assistant",
		Content: []response.ContentBlock{},
	}

	if resp.Text() != "" {
		t.Errorf("got text %q, want empty string", resp.Text())
	}
}

func TestResponse_Text_SkipsToolBlocks(t *testing.T) {
	resp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Let me check that."},
			response.ToolUseBlock{ID: "call_1", Name: "get_weather", Input: map[string]any{"location": "Boston"}},
		},
	}

	if resp.Text() != "Let me check that." {
		t.Errorf("got text %q, want %q", resp.Text(), "Let me check that.")
	}
}

func TestResponse_ToolCalls(t *testing.T) {
	resp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.ToolUseBlock{
				ID:    "call_1",
				Name:  "get_weather",
				Input: map[string]any{"location": "Boston"},
			},
		},
	}

	calls := resp.ToolCalls()
	if len(calls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(calls))
	}

	if calls[0].Name != "get_weather" {
		t.Errorf("got name %q, want %q", calls[0].Name, "get_weather")
	}

	if calls[0].Input["location"] != "Boston" {
		t.Errorf("got location %v, want Boston", calls[0].Input["location"])
	}
}

func TestResponse_ToolCalls_Empty(t *testing.T) {
	resp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Just text."},
		},
	}

	calls := resp.ToolCalls()
	if calls != nil {
		t.Errorf("got %d tool calls, want nil", len(calls))
	}
}

func TestResponse_ToolCalls_MultipleWithText(t *testing.T) {
	resp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "I'll check both."},
			response.ToolUseBlock{ID: "call_1", Name: "get_weather", Input: map[string]any{"location": "Boston"}},
			response.ToolUseBlock{ID: "call_2", Name: "get_weather", Input: map[string]any{"location": "London"}},
		},
	}

	calls := resp.ToolCalls()
	if len(calls) != 2 {
		t.Fatalf("got %d tool calls, want 2", len(calls))
	}

	if calls[0].Input["location"] != "Boston" {
		t.Errorf("got first location %v, want Boston", calls[0].Input["location"])
	}

	if calls[1].Input["location"] != "London" {
		t.Errorf("got second location %v, want London", calls[1].Input["location"])
	}
}

func TestResponse_StopReason(t *testing.T) {
	resp := &response.Response{
		Role:       "assistant",
		StopReason: "stop",
	}

	if resp.StopReason != "stop" {
		t.Errorf("got stop reason %q, want %q", resp.StopReason, "stop")
	}
}

func TestResponse_Usage(t *testing.T) {
	resp := &response.Response{
		Role: "assistant",
		Usage: &response.TokenUsage{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
	}

	if resp.Usage.InputTokens != 10 {
		t.Errorf("got input tokens %d, want 10", resp.Usage.InputTokens)
	}

	if resp.Usage.OutputTokens != 20 {
		t.Errorf("got output tokens %d, want 20", resp.Usage.OutputTokens)
	}

	if resp.Usage.TotalTokens != 30 {
		t.Errorf("got total tokens %d, want 30", resp.Usage.TotalTokens)
	}
}

func TestStreamingResponse_Text(t *testing.T) {
	chunk := &response.StreamingResponse{
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Hello"},
		},
	}

	if chunk.Text() != "Hello" {
		t.Errorf("got text %q, want %q", chunk.Text(), "Hello")
	}
}

func TestStreamingResponse_Text_Empty(t *testing.T) {
	chunk := &response.StreamingResponse{}

	if chunk.Text() != "" {
		t.Errorf("got text %q, want empty string", chunk.Text())
	}
}

func TestEmbeddingsResponse(t *testing.T) {
	resp := &response.EmbeddingsResponse{
		Embeddings: [][]float64{{0.1, 0.2, 0.3}},
		Model:      "text-embedding-ada-002",
		Usage: &response.TokenUsage{
			InputTokens: 5,
			TotalTokens: 5,
		},
	}

	if len(resp.Embeddings) != 1 {
		t.Fatalf("got %d embeddings, want 1", len(resp.Embeddings))
	}

	if len(resp.Embeddings[0]) != 3 {
		t.Errorf("got %d dimensions, want 3", len(resp.Embeddings[0]))
	}

	if resp.Model != "text-embedding-ada-002" {
		t.Errorf("got model %q, want %q", resp.Model, "text-embedding-ada-002")
	}
}
