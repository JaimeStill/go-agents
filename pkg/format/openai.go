package format

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

type openAIChoice struct {
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason,omitempty"`
}

type openAIDelta struct {
	Content string `json:"content"`
}

type openAIEmbedding struct {
	Embedding []float64 `json:"embedding"`
}

type openAIEmbeddingsResponse struct {
	Data  []openAIEmbedding `json:"data"`
	Model string            `json:"model"`
	Usage *openAIUsage      `json:"usage,omitempty"`
}

type openAIMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Usage   *openAIUsage   `json:"usage,omitempty"`
}

type openAIStreamChoice struct {
	Delta        openAIDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type openAIStreamChunk struct {
	Choices []openAIStreamChoice `json:"choices"`
}

type openAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIFactory creates a new OpenAIFormat instance for use with the format registry.
func OpenAIFactory() (Format, error) {
	return &OpenAIFormat{}, nil
}

// OpenAIFormat implements Format for OpenAI-compatible APIs.
type OpenAIFormat struct{}

func (f *OpenAIFormat) Name() string {
	return "openai"
}

func (f *OpenAIFormat) Marshal(proto protocol.Protocol, data any) ([]byte, error) {
	switch proto {
	case protocol.Chat:
		return f.marshalChat(data)
	case protocol.Vision:
		return f.marshalVision(data)
	case protocol.Tools:
		return f.marshalTools(data)
	case protocol.Embeddings:
		return f.marshalEmbeddings(data)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

func (f *OpenAIFormat) Parse(proto protocol.Protocol, body []byte) (any, error) {
	switch proto {
	case protocol.Chat, protocol.Vision:
		return f.parseChat(body)
	case protocol.Tools:
		return f.parseTools(body)
	case protocol.Embeddings:
		return f.parseEmbeddings(body)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

func (f *OpenAIFormat) ParseStreamChunk(proto protocol.Protocol, data []byte) (*response.StreamingResponse, error) {
	if proto == protocol.Embeddings {
		return nil, fmt.Errorf("protocol %s does not support streaming", proto)
	}

	var raw openAIStreamChunk
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse streaming chunk: %w", err)
	}

	chunk := &response.StreamingResponse{}

	if len(raw.Choices) > 0 {
		choice := raw.Choices[0]
		if choice.Delta.Content != "" {
			chunk.Content = append(chunk.Content, response.TextBlock{
				Text: choice.Delta.Content,
			})
		}
		if choice.FinishReason != nil {
			chunk.StopReason = *choice.FinishReason
		}
	}

	return chunk, nil
}

func (f *OpenAIFormat) marshalChat(data any) ([]byte, error) {
	d, ok := data.(*ChatData)
	if !ok {
		return nil, fmt.Errorf("expected *ChatData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["messages"] = d.Messages
	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}

func (f *OpenAIFormat) marshalVision(data any) ([]byte, error) {
	d, ok := data.(*VisionData)
	if !ok {
		return nil, fmt.Errorf("expected *VisionData, got %T", data)
	}

	if len(d.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty for vision requests")
	}

	if len(d.Images) == 0 {
		return nil, fmt.Errorf("images cannot be empty for vision requests")
	}

	lastIdx := len(d.Messages) - 1
	message := d.Messages[lastIdx]

	var textContent string
	switch v := message.Content.(type) {
	case string:
		textContent = v
	default:
		return nil, fmt.Errorf("message content must be a string for vision transformation")
	}

	content := []map[string]any{
		{"type": "text", "text": textContent},
	}

	for _, img := range d.Images {
		var url string
		if img.URL != "" {
			url = img.URL
		} else {
			url = fmt.Sprintf(
				"data:image/%s;base64,%s",
				img.Format,
				base64.StdEncoding.EncodeToString(img.Data),
			)
		}

		imageURL := map[string]any{
			"url": url,
		}

		if d.VisionOptions != nil {
			maps.Copy(imageURL, d.VisionOptions)
		}

		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": imageURL,
		})
	}

	transformedMessages := make([]protocol.Message, len(d.Messages))
	copy(transformedMessages, d.Messages)
	transformedMessages[lastIdx] = protocol.Message{
		Role:    message.Role,
		Content: content,
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["messages"] = transformedMessages
	maps.Copy(combined, d.Options)

	return json.Marshal(combined)
}

func (f *OpenAIFormat) marshalTools(data any) ([]byte, error) {
	d, ok := data.(*ToolsData)
	if !ok {
		return nil, fmt.Errorf("expected *ToolsData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["messages"] = d.Messages

	tools := make([]map[string]any, len(d.Tools))
	for i, tool := range d.Tools {
		tools[i] = map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}

	combined["tools"] = tools

	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}

func (f *OpenAIFormat) marshalEmbeddings(data any) ([]byte, error) {
	d, ok := data.(*EmbeddingsData)
	if !ok {
		return nil, fmt.Errorf("expected *EmbeddingsData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["input"] = d.Input
	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}

func (f *OpenAIFormat) parseChat(body []byte) (*response.Response, error) {
	var raw openAIResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse chat response: %w", err)
	}

	resp := &response.Response{Role: "assistant"}

	if len(raw.Choices) > 0 {
		choice := raw.Choices[0]
		if choice.Message.Content != "" {
			resp.Content = append(resp.Content, response.TextBlock{
				Text: choice.Message.Content,
			})
		}
		resp.StopReason = choice.FinishReason
	}

	resp.Usage = f.mapUsage(raw.Usage)

	return resp, nil
}

func (f *OpenAIFormat) parseTools(body []byte) (*response.Response, error) {
	var raw openAIResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse tools response: %w", err)
	}

	resp := &response.Response{Role: "assistant"}

	if len(raw.Choices) > 0 {
		choice := raw.Choices[0]

		if choice.Message.Content != "" {
			resp.Content = append(resp.Content, response.TextBlock{
				Text: choice.Message.Content,
			})
		}

		for _, tc := range choice.Message.ToolCalls {
			var input map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
				return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
			}

			resp.Content = append(resp.Content, response.ToolUseBlock{
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			})
		}

		resp.StopReason = choice.FinishReason
	}

	resp.Usage = f.mapUsage(raw.Usage)

	return resp, nil
}

func (f *OpenAIFormat) parseEmbeddings(body []byte) (*response.EmbeddingsResponse, error) {
	var raw openAIEmbeddingsResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings response: %w", err)
	}

	embeddings := make([][]float64, len(raw.Data))
	for i, d := range raw.Data {
		embeddings[i] = d.Embedding
	}

	resp := &response.EmbeddingsResponse{
		Embeddings: embeddings,
		Model:      raw.Model,
		Usage:      f.mapUsage(raw.Usage),
	}

	return resp, nil
}

func (f *OpenAIFormat) mapUsage(usage *openAIUsage) *response.TokenUsage {
	if usage == nil {
		return nil
	}

	return &response.TokenUsage{
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		TotalTokens:  usage.TotalTokens,
	}
}
