// Package providers implements LLM service provider integrations.
// It provides a unified Provider interface for interacting with different LLM services
// (Ollama, Azure OpenAI) while handling provider-specific authentication, endpoints,
// and response formats.
//
// # Provider System
//
// The provider system follows a factory pattern with a global registry:
//
//	// Register a provider factory
//	providers.Register("custom", func(c *config.ProviderConfig) (Provider, error) {
//	    // Create and configure provider
//	    return customProvider, nil
//	})
//
//	// Create provider from configuration
//	provider, err := providers.Create(&config.ProviderConfig{
//	    Name:    "ollama",
//	    BaseURL: "http://localhost:11434",
//	    Model:   modelConfig,
//	})
//
// # Provider Interface
//
// All providers implement the Provider interface:
//
//	type Provider interface {
//	    Name() string
//	    BaseURL() string
//
//	    Endpoint(p protocol.Protocol) (string, error)
//	    SetHeaders(ctx context.Context, req *http.Request) error
//
//	    Marshal(proto protocol.Protocol, data any) ([]byte, error)
//	    PrepareRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error)
//	    PrepareStreamRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error)
//	    ProcessResponse(ctx context.Context, resp *http.Response, proto protocol.Protocol) (any, error)
//	    ProcessStreamResponse(ctx context.Context, resp *http.Response, proto protocol.Protocol) (<-chan any, error)
//	}
//
// # Built-in Providers
//
// ## Ollama Provider
//
// Ollama provider connects to local or remote Ollama instances with OpenAI-compatible API:
//
//	cfg := &config.ProviderConfig{
//	    Name:    "ollama",
//	    BaseURL: "http://localhost:11434",
//	    Options: map[string]any{
//	        "auth_type": "bearer",      // Optional: "bearer" or "api_key"
//	        "token":     "your-token",  // Optional: authentication token
//	    },
//	}
//
//	provider, err := providers.NewOllama(cfg)
//
// Features:
//   - Automatic /v1 suffix handling for OpenAI compatibility
//   - Optional bearer or API key authentication
//   - Custom authentication header support
//   - Streaming and non-streaming responses
//
// ## Azure OpenAI Provider
//
// Azure provider integrates with Azure OpenAI Service with deployment-based routing:
//
//	cfg := &config.ProviderConfig{
//	    Name:    "azure",
//	    BaseURL: "https://your-resource.openai.azure.com",
//	    Options: map[string]any{
//	        "deployment":  "gpt-4-deployment",  // Required: deployment name
//	        "auth_type":   "api_key",           // Required: "api_key", "bearer", or "managed_identity"
//	        "token":       "your-api-key",      // Required for api_key/bearer auth types
//	        "api_version": "2024-02-01",        // Required: API version
//	    },
//	}
//
//	provider, err := providers.NewAzure(cfg)
//
// Features:
//   - Deployment-based endpoint routing
//   - API key, Entra ID (bearer token), or managed identity authentication
//   - API version management
//   - Server-sent events with "data: " prefix for streaming
//
// ### Azure Authentication Types
//
// API Key:
//
//	Options: map[string]any{
//	    "auth_type": "api_key",
//	    "token":     "your-api-key",
//	}
//
// Entra ID (Bearer Token):
//
//	Options: map[string]any{
//	    "auth_type": "bearer",
//	    "token":     "your-bearer-token",
//	}
//
// Managed Identity (for containerized deployments):
//
//	Options: map[string]any{
//	    "auth_type": "managed_identity",
//	    "resource":  "https://cognitiveservices.azure.com/.default",  // Optional, this is the default
//	    "client_id": "user-assigned-identity-client-id",             // Optional, for user-assigned identity
//	}
//
// Managed identity acquires tokens dynamically using the Azure SDK's
// ManagedIdentityCredential. Token caching and refresh are handled
// internally by the Azure SDK. No static token is required.
//
// # Base Provider
//
// BaseProvider provides common functionality that provider implementations can embed:
//
//	type CustomProvider struct {
//	    *providers.BaseProvider
//	    // Custom fields
//	}
//
//	func NewCustomProvider(cfg *config.ProviderConfig) (Provider, error) {
//	    return &CustomProvider{
//	        BaseProvider: providers.NewBaseProvider(cfg.Name, cfg.BaseURL),
//	    }, nil
//	}
//
// BaseProvider handles:
//   - Provider name management
//   - Base URL storage
//   - Default OpenAI-compatible request marshaling
//
// # Request and Response Flow
//
// Standard request flow:
//
//	// 1. Get endpoint for protocol
//	endpoint, err := provider.Endpoint(protocol.Chat)
//
//	// 2. Marshal request body
//	body, err := provider.Marshal(protocol.Chat, chatData)
//
//	// 3. Prepare request
//	request, err := provider.PrepareRequest(ctx, protocol.Chat, body, headers)
//
//	// 4. Create HTTP request and set auth headers
//	httpReq, err := http.NewRequestWithContext(ctx, "POST", request.URL, bytes.NewReader(request.Body))
//	for key, value := range request.Headers {
//	    httpReq.Header.Set(key, value)
//	}
//	if err := provider.SetHeaders(ctx, httpReq); err != nil {
//	    return nil, err
//	}
//
//	// 5. Execute request
//	resp, err := httpClient.Do(httpReq)
//
//	// 6. Process response
//	result, err := provider.ProcessResponse(ctx, resp, protocol.Chat)
//
// Streaming request flow:
//
//	// 1-4. Same as standard flow, but use PrepareStreamRequest
//	request, err := provider.PrepareStreamRequest(ctx, protocol.Chat, body, headers)
//
//	// 5. Process streaming response
//	chunks, err := provider.ProcessStreamResponse(ctx, resp, protocol.Chat)
//
//	// 6. Read streaming chunks
//	for chunk := range chunks {
//	    // Handle chunk
//	}
//
// # Request Structure
//
// The Request type packages provider-specific request details:
//
//	type Request struct {
//	    URL     string            // Full endpoint URL
//	    Headers map[string]string // Request headers
//	    Body    []byte            // Marshaled request body
//	}
//
// This structure decouples request preparation from HTTP execution.
//
// # Authentication
//
// Providers handle authentication through the SetHeaders method, which accepts
// a context for providers that need to acquire tokens dynamically (e.g., managed identity):
//
//	// Ollama with bearer token
//	Options: map[string]any{
//	    "auth_type": "bearer",
//	    "token":     "your-token",
//	}
//
//	// Ollama with API key
//	Options: map[string]any{
//	    "auth_type":   "api_key",
//	    "token":       "your-key",
//	    "auth_header": "X-Custom-Auth", // Optional, defaults to "X-API-Key"
//	}
//
//	// Azure with API key
//	Options: map[string]any{
//	    "auth_type": "api_key",
//	    "token":     "your-api-key",
//	}
//
//	// Azure with Entra ID token
//	Options: map[string]any{
//	    "auth_type": "bearer",
//	    "token":     "your-bearer-token",
//	}
//
//	// Azure with managed identity
//	Options: map[string]any{
//	    "auth_type": "managed_identity",
//	}
//
// # Error Handling
//
// Providers return errors for:
//   - Unsupported protocols: Endpoint returns error
//   - Invalid configuration: NewProvider constructors return error
//   - Unsupported auth types: NewAzure returns error
//   - Authentication failures: SetHeaders returns error (e.g., managed identity token acquisition)
//   - HTTP failures: ProcessResponse/ProcessStreamResponse return error with status
//   - Response parsing failures: delegated to response.Parse
//
// # Thread Safety
//
// The provider registry is thread-safe for concurrent registration and creation.
// Individual provider instances are safe for concurrent use after creation.
//
// # Extending with Custom Providers
//
// To implement a custom provider:
//
//  1. Define provider struct (optionally embedding BaseProvider)
//  2. Implement Provider interface methods
//  3. Create factory function: func(c *config.ProviderConfig) (Provider, error)
//  4. Register factory: providers.Register("custom", NewCustomProvider)
//
// Example:
//
//	type CustomProvider struct {
//	    *providers.BaseProvider
//	    apiKey string
//	}
//
//	func NewCustomProvider(cfg *config.ProviderConfig) (providers.Provider, error) {
//	    apiKey, ok := cfg.Options["api_key"].(string)
//	    if !ok || apiKey == "" {
//	        return nil, fmt.Errorf("api_key is required")
//	    }
//
//	    return &CustomProvider{
//	        BaseProvider: providers.NewBaseProvider(cfg.Name, cfg.BaseURL),
//	        apiKey:       apiKey,
//	    }, nil
//	}
//
//	func (p *CustomProvider) Endpoint(proto protocol.Protocol) (string, error) {
//	    // Implement endpoint logic
//	}
//
//	func (p *CustomProvider) SetHeaders(ctx context.Context, req *http.Request) error {
//	    req.Header.Set("Authorization", "Bearer "+p.apiKey)
//	    return nil
//	}
//
//	// Implement remaining Provider interface methods...
package providers
