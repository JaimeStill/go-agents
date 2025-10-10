package agent

import (
	"context"
	"fmt"
	"maps"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/transport"
	"github.com/google/uuid"
)

// Agent provides a high-level interface for LLM interactions.
// Methods are protocol-specific and handle message initialization,
// system prompt injection, and response type assertions.
//
// Each agent has a unique identifier that remains stable across its lifetime.
// The ID is used for orchestration scenarios including hub registration,
// message routing, lifecycle tracking, and distributed tracing.
// IDs are guaranteed to be unique, stable, and thread-safe.
type Agent interface {
	// ID returns the unique identifier for the agent.
	// The ID is assigned at creation time using UUIDv7 and never changes.
	// Thread-safe for concurrent access and safe to use as map keys.
	ID() string

	// Client returns the underlying transport client.
	Client() transport.Client

	// Provider returns the provider instance from the client.
	Provider() providers.Provider

	// Model returns the model instance from the client.
	Model() models.Model

	// Chat executes a chat protocol request with optional system prompt injection.
	// Returns the parsed chat response or an error.
	Chat(ctx context.Context, prompt string, opts ...map[string]any) (*protocols.ChatResponse, error)

	// ChatStream executes a streaming chat protocol request.
	// Automatically sets stream: true in options.
	// Returns a channel of streaming chunks or an error.
	ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan protocols.StreamingChunk, error)

	// Vision executes a vision protocol request with images.
	// Images can be URLs or base64-encoded data URIs.
	// Returns the parsed chat response or an error.
	Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*protocols.ChatResponse, error)

	// VisionStream executes a streaming vision protocol request with images.
	// Returns a channel of streaming chunks or an error.
	VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan protocols.StreamingChunk, error)

	// Tools executes a tools protocol request with function definitions.
	// Returns the parsed tools response with tool calls or an error.
	Tools(ctx context.Context, prompt string, tools []Tool, opts ...map[string]any) (*protocols.ToolsResponse, error)

	// Embed executes an embeddings protocol request.
	// Returns the parsed embeddings response or an error.
	Embed(ctx context.Context, input string, opts ...map[string]any) (*protocols.EmbeddingsResponse, error)
}

// agent implements the Agent interface with transport client orchestration.
type agent struct {
	id           string
	client       transport.Client
	systemPrompt string
}

// New creates a new Agent from configuration.
// Creates the transport client and initializes system prompt.
// Assigns a unique UUIDv7 identifier for orchestration and tracking.
// Returns an error if transport client creation fails.
func New(config *config.AgentConfig) (Agent, error) {
	client, err := transport.New(config.Transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport client: %w", err)
	}

	return &agent{
		id:           uuid.Must(uuid.NewV7()).String(),
		client:       client,
		systemPrompt: config.SystemPrompt,
	}, nil
}

func (a *agent) ID() string {
	return a.id
}

// Client returns the underlying transport client.
func (a *agent) Client() transport.Client {
	return a.client
}

// Provider returns the provider instance from the client.
func (a *agent) Provider() providers.Provider {
	return a.client.Provider()
}

// Model returns the model instance from the client.
func (a *agent) Model() models.Model {
	return a.client.Model()
}

// Chat executes a chat protocol request.
// Initializes messages with system prompt (if configured) and user prompt.
// Merges provided options with model defaults.
// Returns parsed ChatResponse or error.
func (a *agent) Chat(ctx context.Context, prompt string, opts ...map[string]any) (*protocols.ChatResponse, error) {
	messages := a.initMessages(prompt)

	options := make(map[string]any)
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
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

// ChatStream executes a streaming chat protocol request.
// Automatically sets stream: true in options.
// Returns a channel of StreamingChunk or error.
func (a *agent) ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan protocols.StreamingChunk, error) {
	messages := a.initMessages(prompt)

	options := make(map[string]any)
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	options["stream"] = true

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: messages,
		Options:  options,
	}

	return a.client.ExecuteProtocolStream(ctx, req)
}

// Vision executes a vision protocol request with images.
// Images can be URLs or base64-encoded data URIs.
// Converts images to []any format for capability processing.
// Returns parsed ChatResponse or error.
func (a *agent) Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*protocols.ChatResponse, error) {
	messages := a.initMessages(prompt)

	// Convert []string to []any for capability processing
	imageList := make([]any, len(images))
	for i, img := range images {
		imageList[i] = img
	}

	options := map[string]any{
		"images": imageList,
	}

	if len(opts) > 0 && opts[0] != nil {
		maps.Copy(options, opts[0])
	}

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

// VisionStream executes a streaming vision protocol request with images.
// Automatically sets stream: true in options.
// Returns a channel of StreamingChunk or error.
func (a *agent) VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan protocols.StreamingChunk, error) {
	messages := a.initMessages(prompt)

	// Convert []string to []any for capability processing
	imageList := make([]any, len(images))
	for i, img := range images {
		imageList[i] = img
	}

	options := map[string]any{
		"images": imageList,
	}

	if len(opts) > 0 && opts[0] != nil {
		maps.Copy(options, opts[0])
	}

	options["stream"] = true

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Vision,
		Messages: messages,
		Options:  options,
	}

	return a.client.ExecuteProtocolStream(ctx, req)
}

// Tools executes a tools protocol request with function definitions.
// Converts Tool structs to FunctionDefinition format.
// Returns parsed ToolsResponse with tool calls or error.
func (a *agent) Tools(ctx context.Context, prompt string, tools []Tool, opts ...map[string]any) (*protocols.ToolsResponse, error) {
	messages := a.initMessages(prompt)

	options := map[string]any{
		"tools": setToolDefinitions(tools),
	}

	if len(opts) > 0 && opts[0] != nil {
		maps.Copy(options, opts[0])
	}

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Tools,
		Messages: messages,
		Options:  options,
	}

	result, err := a.client.ExecuteProtocol(ctx, req)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*protocols.ToolsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	return response, nil
}

// Embed executes an embeddings protocol request.
// Sets input text in options and sends empty message list.
// Returns parsed EmbeddingsResponse or error.
func (a *agent) Embed(ctx context.Context, input string, opts ...map[string]any) (*protocols.EmbeddingsResponse, error) {
	options := map[string]any{
		"input": input,
	}

	if len(opts) > 0 && opts[0] != nil {
		maps.Copy(options, opts[0])
	}

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

// setToolDefinitions converts Tool structs to FunctionDefinition format.
// Each tool is wrapped in a function definition with type "function".
func setToolDefinitions(tools []Tool) []capabilities.FunctionDefinition {
	defs := make([]capabilities.FunctionDefinition, len(tools))
	for i, tool := range tools {
		defs[i] = capabilities.FunctionDefinition{
			Type: "function",
			Function: map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}

	return defs
}

// initMessages creates the initial message list with optional system prompt.
// If system prompt is configured, it's added as the first message.
// User prompt is always added after system prompt.
func (a *agent) initMessages(prompt string) []protocols.Message {
	messages := make([]protocols.Message, 0)

	if a.systemPrompt != "" {
		messages = append(messages, protocols.NewMessage("system", a.systemPrompt))
	}

	messages = append(messages, protocols.NewMessage("user", prompt))

	return messages
}

// Tool defines a function that can be called by the LLM.
// Used with the Tools protocol for function calling capabilities.
type Tool struct {
	// Name is the function name that the LLM will call.
	Name string `json:"name"`

	// Description explains what the function does.
	// Should be clear and detailed to help the LLM decide when to use it.
	Description string `json:"description"`

	// Parameters is a JSON Schema defining the function's parameters.
	// Uses the format: {"type": "object", "properties": {...}, "required": [...]}
	Parameters map[string]any `json:"parameters"`
}
