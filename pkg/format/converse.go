package format

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

type converseContentBlock struct {
	Text    string           `json:"text,omitempty"`
	ToolUse *converseToolUse `json:"toolUse,omitempty"`
}

type converseContentBlockDelta struct {
	ContentBlockIndex int           `json:"contentBlockIndex"`
	Delta             converseDelta `json:"delta"`
}

type converseDelta struct {
	Text string `json:"text"`
}

type converseMessage struct {
	Role    string                 `json:"role"`
	Content []converseContentBlock `json:"content"`
}

type converseMessageStop struct {
	StopReason string `json:"stopReason"`
}

type converseOutput struct {
	Message converseMessage `json:"message"`
}

type converseResponse struct {
	Output     converseOutput `json:"output"`
	StopReason string         `json:"stopReason"`
	Usage      converseUsage  `json:"usage"`
}

type converseStreamEvent struct {
	ContentBlockDelta *converseContentBlockDelta `json:"contentBlockDelta,omitempty"`
	MessageStop       *converseMessageStop
}

type converseToolUse struct {
	ToolUseID string         `json:"toolUseId"`
	Name      string         `json:"name"`
	Input     map[string]any `json:"input"`
}

type converseUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// ConverseFormat implements Format for the AWS Bedrock Converse API.
type ConverseFormat struct{}

func (f *ConverseFormat) Name() string {
	return "converse"
}

// ConverseFactory creates a new ConverseFormat instance for use with the format registry.
func ConverseFactory() (Format, error) {
	return &ConverseFormat{}, nil
}

func (f *ConverseFormat) Marshal(proto protocol.Protocol, data any) ([]byte, error) {
	switch proto {
	case protocol.Chat:
		return f.marshalChat(data)
	case protocol.Vision:
		return f.marshalVision(data)
	case protocol.Tools:
		return f.marshalTools(data)
	case protocol.Embeddings:
		return nil, fmt.Errorf("embeddings protocol is not supported by Converse API")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

func (f *ConverseFormat) Parse(proto protocol.Protocol, body []byte) (any, error) {
	switch proto {
	case protocol.Chat, protocol.Vision, protocol.Tools:
		return f.parseResponse(body)
	case protocol.Embeddings:
		return nil, fmt.Errorf("embeddings protocol not supported by Converse API")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

func (f *ConverseFormat) ParseStreamChunk(proto protocol.Protocol, data []byte) (*response.StreamingResponse, error) {
	var event converseStreamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to parse stream event: %w", err)
	}

	if event.ContentBlockDelta != nil {
		return &response.StreamingResponse{
			Content: []response.ContentBlock{
				response.TextBlock{Text: event.ContentBlockDelta.Delta.Text},
			},
		}, nil
	}

	if event.MessageStop != nil {
		return &response.StreamingResponse{
			StopReason: event.MessageStop.StopReason,
		}, nil
	}

	return nil, nil
}

func (f *ConverseFormat) buildImageBlock(img Image) map[string]any {
	return map[string]any{
		"image": map[string]any{
			"format": img.Format,
			"source": map[string]any{
				"bytes": img.Data,
			},
		},
	}
}

func (f *ConverseFormat) marshalChat(data any) ([]byte, error) {
	d, ok := data.(*ChatData)
	if !ok {
		return nil, fmt.Errorf("expected *ChatData, got %T", data)
	}

	req := make(map[string]any)
	req["modelId"] = d.Model

	system, messages := f.splitSystemMessages(d.Messages)
	if len(system) > 0 {
		req["system"] = system
	}
	req["messages"] = messages

	maps.Copy(req, d.Options)

	return json.Marshal(req)
}

func (f *ConverseFormat) marshalVision(data any) ([]byte, error) {
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

	req := make(map[string]any)
	req["modelId"] = d.Model

	system, messages := f.splitSystemMessages(d.Messages)
	if len(system) > 0 {
		req["system"] = system
	}

	lastIdx := len(messages) - 1
	lastMsg := messages[lastIdx].(map[string]any)
	content := lastMsg["content"].([]map[string]any)

	for _, img := range d.Images {
		content = append(content, f.buildImageBlock(img))
	}
	lastMsg["content"] = content
	messages[lastIdx] = lastMsg
	req["messages"] = messages

	maps.Copy(req, d.Options)

	return json.Marshal(req)
}

func (f *ConverseFormat) marshalTools(data any) ([]byte, error) {
	d, ok := data.(*ToolsData)
	if !ok {
		return nil, fmt.Errorf("expected *ToolsData, got %T", data)
	}

	req := make(map[string]any)
	req["modelId"] = d.Model

	system, messages := f.splitSystemMessages(d.Messages)
	if len(system) > 0 {
		req["system"] = system
	}
	req["messages"] = messages

	tools := make([]map[string]any, len(d.Tools))
	for i, tool := range d.Tools {
		tools[i] = map[string]any{
			"toolSpec": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": map[string]any{
					"json": tool.Parameters,
				},
			},
		}
	}

	req["toolConfig"] = map[string]any{
		"tools": tools,
	}

	maps.Copy(req, d.Options)

	return json.Marshal(req)
}

func (f *ConverseFormat) parseResponse(body []byte) (*response.Response, error) {
	var raw converseResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse Converse response: %w", err)
	}

	resp := &response.Response{
		Role:       raw.Output.Message.Role,
		StopReason: raw.StopReason,
		Usage: &response.TokenUsage{
			InputTokens:  raw.Usage.InputTokens,
			OutputTokens: raw.Usage.OutputTokens,
			TotalTokens:  raw.Usage.TotalTokens,
		},
	}

	for _, block := range raw.Output.Message.Content {
		if block.Text != "" {
			resp.Content = append(resp.Content, response.TextBlock{
				Text: block.Text,
			})
		}
		if block.ToolUse != nil {
			resp.Content = append(resp.Content, response.ToolUseBlock{
				ID:    block.ToolUse.ToolUseID,
				Name:  block.ToolUse.Name,
				Input: block.ToolUse.Input,
			})
		}
	}

	return resp, nil
}

func (f *ConverseFormat) splitSystemMessages(msgs []protocol.Message) ([]map[string]any, []any) {
	var system []map[string]any
	var messages []any

	for _, msg := range msgs {
		if msg.Role == "system" {
			if text, ok := msg.Content.(string); ok {
				system = append(system, map[string]any{"text": text})
			}
			continue
		}

		content := []map[string]any{}
		if text, ok := msg.Content.(string); ok {
			content = append(content, map[string]any{"text": text})
		}

		messages = append(messages, map[string]any{
			"role":    msg.Role,
			"content": content,
		})
	}

	return system, messages
}
