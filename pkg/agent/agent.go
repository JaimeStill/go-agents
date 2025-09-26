package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/transport"
)

type Agent interface {
	Client() transport.Client
	Provider() providers.Provider
	Model() models.Model

	Chat(ctx context.Context, prompt string) (*protocols.ChatResponse, error)
	ChatStream(ctx context.Context, prompt string) (<-chan protocols.StreamingChunk, error)

	Vision(ctx context.Context, prompt string, images []string) (*protocols.ChatResponse, error)
	VisionStream(ctx context.Context, prompt string, images []string) (<-chan protocols.StreamingChunk, error)

	Tools(ctx context.Context, prompt string, tools []Tool) (*protocols.ChatResponse, error)
	ToolsStream(ctx context.Context, prompt string, tools []Tool) (<-chan protocols.StreamingChunk, error)

	Embed(ctx context.Context, input string) (*protocols.EmbeddingsResponse, error)
}

type agent struct {
	client       transport.Client
	systemPrompt string
	maxRetries   int
	timeout      time.Duration
}

func New(config *config.AgentConfig) (Agent, error) {
	client, err := transport.New(config.Transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport client: %w", err)
	}

	return &agent{
		client:       client,
		systemPrompt: config.SystemPrompt,
	}, nil
}

func (a *agent) Client() transport.Client {
	return a.client
}

func (a *agent) Provider() providers.Provider {
	return a.client.Provider()
}

func (a *agent) Model() models.Model {
	return a.client.Model()
}

func (a *agent) Chat(ctx context.Context, prompt string) (*protocols.ChatResponse, error) {
	messages := a.initMessages(prompt)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: messages,
		Options:  a.Model().Options(),
	}

	result, err := a.client.ExecuteProtocol(ctx, req)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*protocols.ChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	return response, nil
}

func (a *agent) ChatStream(ctx context.Context, prompt string) (<-chan protocols.StreamingChunk, error) {
	messages := a.initMessages(prompt)

	options := a.Model().Options()
	options["stream"] = true

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: messages,
		Options:  options,
	}

	return a.client.ExecuteProtocolStream(ctx, req)
}

func (a *agent) Vision(ctx context.Context, prompt string, images []string) (*protocols.ChatResponse, error) {
	messages := a.initMessages(prompt)

	// Convert []string to []any for capability processing
	imageList := make([]any, len(images))
	for i, img := range images {
		imageList[i] = img
	}

	options := a.Model().Options()
	options["images"] = imageList

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Vision,
		Messages: messages,
		Options:  options,
	}

	result, err := a.client.ExecuteProtocol(ctx, req)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*protocols.ChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	return response, nil
}

func (a *agent) VisionStream(ctx context.Context, prompt string, images []string) (<-chan protocols.StreamingChunk, error) {
	messages := a.initMessages(prompt)

	// Convert []string to []any for capability processing
	imageList := make([]any, len(images))
	for i, img := range images {
		imageList[i] = img
	}

	options := a.Model().Options()
	options["images"] = imageList
	options["stream"] = true

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Vision,
		Messages: messages,
		Options:  options,
	}

	return a.client.ExecuteProtocolStream(ctx, req)
}

func (a *agent) Tools(ctx context.Context, prompt string, tools []Tool) (*protocols.ChatResponse, error) {
	messages := a.initMessages(prompt)

	options := a.Model().Options()
	options["tools"] = setToolDefinitions(tools)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Tools,
		Messages: messages,
		Options:  options,
	}

	result, err := a.client.ExecuteProtocol(ctx, req)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*protocols.ChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	return response, nil
}

func (a *agent) ToolsStream(ctx context.Context, prompt string, tools []Tool) (<-chan protocols.StreamingChunk, error) {
	messages := a.initMessages(prompt)

	options := a.Model().Options()
	options["tools"] = setToolDefinitions(tools)
	options["stream"] = true

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Tools,
		Messages: messages,
		Options:  options,
	}

	return a.client.ExecuteProtocolStream(ctx, req)
}

func (a *agent) Embed(ctx context.Context, input string) (*protocols.EmbeddingsResponse, error) {
	options := a.Model().Options()
	options["input"] = input

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Embeddings,
		Messages: []protocols.Message{},
		Options:  options,
	}

	result, err := a.client.ExecuteProtocol(ctx, req)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*protocols.EmbeddingsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	return response, nil
}

func setToolDefinitions(tools []Tool) []map[string]any {
	defs := make([]map[string]any, len(tools))
	for i, tool := range tools {
		defs[i] = map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}

	return defs
}

func (a *agent) initMessages(prompt string) []protocols.Message {
	messages := make([]protocols.Message, 0)

	if a.systemPrompt != "" {
		messages = append(messages, protocols.NewMessage("system", a.systemPrompt))
	}

	messages = append(messages, protocols.NewMessage("user", prompt))

	return messages
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}
