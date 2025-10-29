package agent

import (
	"context"
	"fmt"
	"maps"

	"github.com/JaimeStill/go-agents/pkg/client"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/types"
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

	// Client returns the underlying client.
	Client() client.Client

	// Provider returns the provider instance from the client.
	Provider() providers.Provider

	// Model returns the model instance from the client.
	Model() *types.Model

	// Chat executes a chat protocol request with optional system prompt injection.
	// Returns the parsed chat response or an error.
	Chat(ctx context.Context, prompt string, opts ...map[string]any) (*types.ChatResponse, error)

	// ChatStream executes a streaming chat protocol request.
	// Automatically sets stream: true in options.
	// Returns a channel of streaming chunks or an error.
	ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *types.StreamingChunk, error)

	// Vision executes a vision protocol request with images.
	// Images can be URLs or base64-encoded data URIs.
	// Returns the parsed chat response or an error.
	Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*types.ChatResponse, error)

	// VisionStream executes a streaming vision protocol request with images.
	// Returns a channel of streaming chunks or an error.
	VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan *types.StreamingChunk, error)

	// Tools executes a tools protocol request with function definitions.
	// Returns the parsed tools response with tool calls or an error.
	Tools(ctx context.Context, prompt string, tools []Tool, opts ...map[string]any) (*types.ToolsResponse, error)

	// Embed executes an embeddings protocol request.
	// Returns the parsed embeddings response or an error.
	Embed(ctx context.Context, input string, opts ...map[string]any) (*types.EmbeddingsResponse, error)
}

// agent implements the Agent interface with client orchestration.
type agent struct {
	id           string
	client       client.Client
	systemPrompt string
}

// New creates a new Agent from configuration.
// Creates the client and initializes system prompt.
// Assigns a unique UUIDv7 identifier for orchestration and tracking.
// Returns an error if client creation fails.
func New(config *config.AgentConfig) (Agent, error) {
	c, err := client.New(config.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &agent{
		id:           uuid.Must(uuid.NewV7()).String(),
		client:       c,
		systemPrompt: config.SystemPrompt,
	}, nil
}

func (a *agent) ID() string {
	return a.id
}

// Client returns the underlying client.
func (a *agent) Client() client.Client {
	return a.client
}

// Provider returns the provider instance from the client.
func (a *agent) Provider() providers.Provider {
	return a.client.Provider()
}

// Model returns the model instance from the client.
func (a *agent) Model() *types.Model {
	return a.client.Model()
}

// Chat executes a chat protocol request.
// Initializes messages with system prompt (if configured) and user prompt.
// Creates a ChatRequest and passes it to the client.
// Returns parsed ChatResponse or error.
func (a *agent) Chat(ctx context.Context, prompt string, opts ...map[string]any) (*types.ChatResponse, error) {
	messages := a.initMessages(prompt)

	options := make(map[string]any)
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	// Add model name to options
	options["model"] = a.client.Model().Name

	request := &types.ChatRequest{
		Messages: messages,
		Options:  options,
	}

	result, err := a.client.ExecuteProtocol(ctx, request)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*types.ChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return response, nil
}

// ChatStream executes a streaming chat protocol request.
// Automatically sets stream: true in options.
// Returns a channel of StreamingChunk or error.
func (a *agent) ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *types.StreamingChunk, error) {
	messages := a.initMessages(prompt)

	options := make(map[string]any)
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	// Add model name and streaming flag to options
	options["model"] = a.client.Model().Name
	options["stream"] = true

	request := &types.ChatRequest{
		Messages: messages,
		Options:  options,
	}

	return a.client.ExecuteProtocolStream(ctx, request)
}

// Vision executes a vision protocol request with images.
// Images can be URLs or base64-encoded data URIs.
// Extracts image_options from opts if present, separating them from model options.
// Returns parsed ChatResponse or error.
func (a *agent) Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*types.ChatResponse, error) {
	messages := a.initMessages(prompt)

	options := make(map[string]any)
	var imageOptions map[string]any

	if len(opts) > 0 && opts[0] != nil {
		// Extract image_options if present
		if imgOpts, exists := opts[0]["image_options"]; exists {
			if imgOptsMap, ok := imgOpts.(map[string]any); ok {
				imageOptions = imgOptsMap
			}
			// Copy all options except image_options
			for k, v := range opts[0] {
				if k != "image_options" {
					options[k] = v
				}
			}
		} else {
			maps.Copy(options, opts[0])
		}
	}

	// Add model name to options
	options["model"] = a.client.Model().Name

	request := &types.VisionRequest{
		Messages:     messages,
		Images:       images,
		ImageOptions: imageOptions,
		Options:      options,
	}

	result, err := a.client.ExecuteProtocol(ctx, request)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*types.ChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return response, nil
}

// VisionStream executes a streaming vision protocol request with images.
// Automatically sets stream: true in options.
// Returns a channel of StreamingChunk or error.
func (a *agent) VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan *types.StreamingChunk, error) {
	messages := a.initMessages(prompt)

	options := make(map[string]any)
	var imageOptions map[string]any

	if len(opts) > 0 && opts[0] != nil {
		// Extract image_options if present
		if imgOpts, exists := opts[0]["image_options"]; exists {
			if imgOptsMap, ok := imgOpts.(map[string]any); ok {
				imageOptions = imgOptsMap
			}
			// Copy all options except image_options
			for k, v := range opts[0] {
				if k != "image_options" {
					options[k] = v
				}
			}
		} else {
			maps.Copy(options, opts[0])
		}
	}

	// Add model name and streaming flag to options
	options["model"] = a.client.Model().Name
	options["stream"] = true

	request := &types.VisionRequest{
		Messages:     messages,
		Images:       images,
		ImageOptions: imageOptions,
		Options:      options,
	}

	return a.client.ExecuteProtocolStream(ctx, request)
}

// Tools executes a tools protocol request with function definitions.
// Converts agent.Tool structs to types.ToolDefinition format.
// Returns parsed ToolsResponse with tool calls or error.
func (a *agent) Tools(ctx context.Context, prompt string, tools []Tool, opts ...map[string]any) (*types.ToolsResponse, error) {
	messages := a.initMessages(prompt)

	options := make(map[string]any)
	if len(opts) > 0 && opts[0] != nil {
		maps.Copy(options, opts[0])
	}

	// Add model name to options
	options["model"] = a.client.Model().Name

	// Convert agent.Tool to types.ToolDefinition
	toolDefs := make([]types.ToolDefinition, len(tools))
	for i, tool := range tools {
		toolDefs[i] = types.ToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		}
	}

	request := &types.ToolsRequest{
		Messages: messages,
		Tools:    toolDefs,
		Options:  options,
	}

	result, err := a.client.ExecuteProtocol(ctx, request)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*types.ToolsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return response, nil
}

// Embed executes an embeddings protocol request.
// Creates an EmbeddingsRequest with input text and options.
// Returns parsed EmbeddingsResponse or error.
func (a *agent) Embed(ctx context.Context, input string, opts ...map[string]any) (*types.EmbeddingsResponse, error) {
	options := make(map[string]any)
	if len(opts) > 0 && opts[0] != nil {
		maps.Copy(options, opts[0])
	}

	// Add model name to options
	options["model"] = a.client.Model().Name

	request := &types.EmbeddingsRequest{
		Input:   input,
		Options: options,
	}

	result, err := a.client.ExecuteProtocol(ctx, request)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*types.EmbeddingsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", result)
	}

	return response, nil
}

// FunctionDefinition represents a tool definition for the Tools protocol.
// Used internally to convert Agent Tool structs to protocol format.
type FunctionDefinition struct {
	Type     string         `json:"type"`
	Function map[string]any `json:"function"`
}

// setToolDefinitions converts Tool structs to FunctionDefinition format.
// Each tool is wrapped in a function definition with type "function".
func setToolDefinitions(tools []Tool) []FunctionDefinition {
	defs := make([]FunctionDefinition, len(tools))
	for i, tool := range tools {
		defs[i] = FunctionDefinition{
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
func (a *agent) initMessages(prompt string) []types.Message {
	messages := make([]types.Message, 0)

	if a.systemPrompt != "" {
		messages = append(messages, types.NewMessage("system", a.systemPrompt))
	}

	messages = append(messages, types.NewMessage("user", prompt))

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
