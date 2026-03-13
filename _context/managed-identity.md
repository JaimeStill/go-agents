# Azure Managed Identity Support

This document details the changes required to support Azure Managed Identity authentication in the Azure provider. The implementation is structured in two phases: a preparation phase that updates the `SetHeaders` interface signature across all providers, and a feature phase that adds managed identity token acquisition via the `azidentity` SDK.

## Problem Context

The Azure provider unconditionally requires a static `token` at provider creation time. For containerized apps using Azure Managed Identity (e.g., Azure Container Apps), tokens must be acquired dynamically per-request via Azure's credential chain. This blocks deploying go-agents-based apps with managed identity authentication.

## Architecture Approach

A new `pkg/identities/` package isolates the `azidentity` SDK dependency below `pkg/providers/` in the dependency hierarchy. The Azure provider depends on `pkg/identities` for managed identity tokens, but no Azure SDK types leak into the `Provider` interface or any other package.

**Updated dependency hierarchy:**

```
pkg/config          (foundation)
pkg/protocol        (protocol types)
pkg/response        (response parsing)
pkg/identities      (managed identity token sources) ← NEW
pkg/providers       (provider implementations, depends on identities)
pkg/model           (model runtime)
pkg/request         (request interface)
pkg/client          (HTTP orchestration)
pkg/agent           (high-level)
pkg/mock            (testing)
```

## Preparation Phase: SetHeaders Interface Change

This phase updates the `Provider.SetHeaders` signature from `SetHeaders(req *http.Request)` to `SetHeaders(ctx context.Context, req *http.Request) error`. This is required because token acquisition performs I/O (needs context for cancellation) and can fail (needs error return).

### Step 1: Update Provider Interface

**File: `pkg/providers/provider.go`**

**Current code (line 28-30):**
```go
	// SetHeaders sets provider-specific authentication and custom headers on an HTTP request.
	// This is called after the request is created but before it is executed.
	SetHeaders(req *http.Request)
```

**Updated code:**
```go
	// SetHeaders sets provider-specific authentication and custom headers on an HTTP request.
	// This is called after the request is created but before it is executed.
	// Accepts context for operations that require I/O (e.g., managed identity token acquisition).
	// Returns an error if authentication setup fails.
	SetHeaders(ctx context.Context, req *http.Request) error
```

Note: `context` is already imported in this file.

### Step 2: Update Ollama Provider

**File: `pkg/providers/ollama.go`**

**Current code (line 176-194):**
```go
// SetHeaders sets authentication headers on the HTTP request.
// Supports "bearer" token (Authorization: Bearer <token>) and "api_key" (custom header).
// The "auth_header" option allows customizing the API key header name (default: X-API-Key).
func (p *OllamaProvider) SetHeaders(req *http.Request) {
	if authType, ok := p.options["auth_type"].(string); ok {
		if token, ok := p.options["token"].(string); ok && token != "" {
			switch authType {
			case "bearer":
				req.Header.Set("Authorization", "Bearer "+token)
			case "api_key":
				headerName := "X-API-Key"
				if head, ok := p.options["auth_header"].(string); ok && head != "" {
					headerName = head
				}
				req.Header.Set(headerName, token)
			}
		}
	}
}
```

**Updated code:**
```go
// SetHeaders sets authentication headers on the HTTP request.
// Supports "bearer" token (Authorization: Bearer <token>) and "api_key" (custom header).
// The "auth_header" option allows customizing the API key header name (default: X-API-Key).
func (p *OllamaProvider) SetHeaders(ctx context.Context, req *http.Request) error {
	if authType, ok := p.options["auth_type"].(string); ok {
		if token, ok := p.options["token"].(string); ok && token != "" {
			switch authType {
			case "bearer":
				req.Header.Set("Authorization", "Bearer "+token)
			case "api_key":
				headerName := "X-API-Key"
				if head, ok := p.options["auth_header"].(string); ok && head != "" {
					headerName = head
				}
				req.Header.Set(headerName, token)
			}
		}
	}

	return nil
}
```

### Step 3: Update Azure Provider SetHeaders

**File: `pkg/providers/azure.go`**

**Current code (line 202-215):**
```go
// SetHeaders sets authentication headers on the HTTP request.
// Supports "api_key" (api-key header) and "bearer" (Authorization: Bearer <token>).
func (p *AzureProvider) SetHeaders(req *http.Request) {
	switch p.authType {
	case "api_key":
		if p.token != "" {
			req.Header.Set("api-key", p.token)
		}
	case "bearer":
		if p.token != "" {
			req.Header.Set("Authorization", "Bearer "+p.token)
		}
	}
}
```

**Updated code:**
```go
// SetHeaders sets authentication headers on the HTTP request.
// Supports "api_key" (api-key header), "bearer" (Authorization: Bearer <token>),
// and "managed_identity" (dynamic token acquisition via Azure identity).
func (p *AzureProvider) SetHeaders(ctx context.Context, req *http.Request) error {
	switch p.authType {
	case "api_key":
		if p.token != "" {
			req.Header.Set("api-key", p.token)
		}
	case "bearer":
		if p.token != "" {
			req.Header.Set("Authorization", "Bearer "+p.token)
		}
	}

	return nil
}
```

Note: The `managed_identity` case will be added in the Feature Phase after `pkg/identities` exists.

### Step 4: Update Mock Provider

**File: `pkg/mock/provider.go`**

**Add field to MockProvider struct (after line 30, the `endpointError` field):**
```go
	setHeadersError       error
```

**Add option function (after WithEndpointError, around line 126):**
```go
// WithSetHeadersError sets an error for SetHeaders.
func WithSetHeadersError(err error) MockProviderOption {
	return func(m *MockProvider) {
		m.setHeadersError = err
	}
}
```

**Current code (line 153-158):**
```go
// SetHeaders sets the configured headers on the request.
func (m *MockProvider) SetHeaders(req *http.Request) {
	for key, value := range m.headers {
		req.Header.Set(key, value)
	}
}
```

**Updated code:**
```go
// SetHeaders sets the configured headers on the request.
func (m *MockProvider) SetHeaders(ctx context.Context, req *http.Request) error {
	if m.setHeadersError != nil {
		return m.setHeadersError
	}

	for key, value := range m.headers {
		req.Header.Set(key, value)
	}

	return nil
}
```

Note: Add `"context"` to the import block.

### Step 5: Update Client Call Sites

**File: `pkg/client/client.go`**

**Current code (line 117):**
```go
	provider.SetHeaders(httpReq)
```

**Updated code:**
```go
	if err := provider.SetHeaders(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("failed to set auth headers: %w", err)
	}
```

**Current code (line 197):**
```go
	provider.SetHeaders(httpReq)
```

**Updated code:**
```go
	if err := provider.SetHeaders(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("failed to set auth headers: %w", err)
	}
```

### Step 5 Verification

At this point, run:

```bash
go build ./...
```

All packages should compile. Existing tests should pass (with signature updates handled by the compiler — tests that call `SetHeaders` directly will need the new signature, but the provider tests don't call `SetHeaders` directly).

```bash
go test ./tests/...
```

## Feature Phase: Managed Identity Implementation

### Step 6: Add Azure SDK Dependencies

```bash
go get github.com/Azure/azure-sdk-for-go/sdk/azidentity
go mod tidy
```

This pulls in `azidentity` and its transitive dependency `azcore`.

### Step 7: Create pkg/identities/azure.go

**New file: `pkg/identities/azure.go`**

```go
package identities

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const defaultScope = "https://cognitiveservices.azure.com/.default"

// AzureTokenSource acquires bearer tokens using Azure managed identity.
// Uses ManagedIdentityCredential for direct managed identity authentication.
// Token caching is handled internally by the Azure SDK.
type AzureTokenSource struct {
	cred  azcore.TokenCredential
	scope string
}

// NewAzureTokenSource creates a new AzureTokenSource.
// scope defaults to "https://cognitiveservices.azure.com/.default" if empty.
// clientID is optional — when provided, it configures user-assigned managed identity
// via ManagedIdentityCredentialOptions.ID. When empty, system-assigned identity is used.
func NewAzureTokenSource(scope, clientID string) (*AzureTokenSource, error) {
	if scope == "" {
		scope = defaultScope
	}

	opts := &azidentity.ManagedIdentityCredentialOptions{}

	if clientID != "" {
		opts.ID = azidentity.ClientID(clientID)
	}

	cred, err := azidentity.NewManagedIdentityCredential(opts)
	if err != nil {
		return nil, fmt.Errorf("create managed identity credential: %w", err)
	}

	return &AzureTokenSource{
		cred:  cred,
		scope: scope,
	}, nil
}

// GetToken acquires a bearer token for the configured scope.
// The Azure SDK handles token caching and refresh internally.
func (s *AzureTokenSource) GetToken(ctx context.Context) (string, error) {
	token, err := s.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{s.scope},
	})
	if err != nil {
		return "", fmt.Errorf("acquire token for scope %q: %w", s.scope, err)
	}

	return token.Token, nil
}
```

### Step 8: Integrate Managed Identity into Azure Provider

**File: `pkg/providers/azure.go`**

**Add import:**
```go
	"github.com/JaimeStill/go-agents/pkg/identities"
```

**Current struct (line 19-25):**
```go
type AzureProvider struct {
	*BaseProvider
	deployment string
	authType   string
	token      string
	apiVersion string
}
```

**Updated struct:**
```go
type AzureProvider struct {
	*BaseProvider
	deployment  string
	authType    string
	token       string
	apiVersion  string
	tokenSource *identities.AzureTokenSource
}
```

**Current NewAzure (line 30-58):**
```go
func NewAzure(c *config.ProviderConfig) (Provider, error) {
	deployment, ok := c.Options["deployment"].(string)
	if !ok || deployment == "" {
		return nil, fmt.Errorf("deployment is required for Azure provider")
	}

	authType, ok := c.Options["auth_type"].(string)
	if !ok || authType == "" {
		return nil, fmt.Errorf("auth_type is required for Azure provider")
	}

	token, ok := c.Options["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("token is required for Azure provider")
	}

	apiVersion, ok := c.Options["api_version"].(string)
	if !ok || apiVersion == "" {
		return nil, fmt.Errorf("api_version is required for Azure provider")
	}

	return &AzureProvider{
		BaseProvider: NewBaseProvider(c.Name, c.BaseURL),
		deployment:   deployment,
		authType:     authType,
		token:        token,
		apiVersion:   apiVersion,
	}, nil
}
```

**Updated NewAzure:**
```go
func NewAzure(c *config.ProviderConfig) (Provider, error) {
	deployment, ok := c.Options["deployment"].(string)
	if !ok || deployment == "" {
		return nil, fmt.Errorf("deployment is required for Azure provider")
	}

	authType, ok := c.Options["auth_type"].(string)
	if !ok || authType == "" {
		return nil, fmt.Errorf("auth_type is required for Azure provider")
	}

	apiVersion, ok := c.Options["api_version"].(string)
	if !ok || apiVersion == "" {
		return nil, fmt.Errorf("api_version is required for Azure provider")
	}

	p := &AzureProvider{
		BaseProvider: NewBaseProvider(c.Name, c.BaseURL),
		deployment:   deployment,
		authType:     authType,
		apiVersion:   apiVersion,
	}

	if err := p.initAuth(c); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *AzureProvider) initAuth(c *config.ProviderConfig) error {
	switch p.authType {
	case "api_key", "bearer":
		token, ok := c.Options["token"].(string)
		if !ok || token == "" {
			return fmt.Errorf("token is required for Azure provider with auth_type %q", p.authType)
		}
		p.token = token
	case "managed_identity":
		resource, _ := c.Options["resource"].(string)
		clientID, _ := c.Options["client_id"].(string)

		tokenSource, err := identities.NewAzureTokenSource(resource, clientID)
		if err != nil {
			return fmt.Errorf("initialize managed identity: %w", err)
		}
		p.tokenSource = tokenSource
	default:
		return fmt.Errorf("unsupported auth_type %q for Azure provider", p.authType)
	}

	return nil
}
```

**Update SetHeaders (from Step 3) to add managed_identity case:**

Replace the SetHeaders written in Step 3 with:

```go
func (p *AzureProvider) SetHeaders(ctx context.Context, req *http.Request) error {
	switch p.authType {
	case "api_key":
		if p.token != "" {
			req.Header.Set("api-key", p.token)
		}
	case "bearer":
		if p.token != "" {
			req.Header.Set("Authorization", "Bearer "+p.token)
		}
	case "managed_identity":
		token, err := p.tokenSource.GetToken(ctx)
		if err != nil {
			return fmt.Errorf("acquire managed identity token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return nil
}
```

### Step 8 Verification

```bash
go build ./...
go test ./tests/...
```

All packages should compile and existing tests should pass. The `TestNewAzure_MissingToken` test will still pass because it uses `auth_type: "api_key"`, which still requires a token.

## Configuration Examples

### API Key (unchanged)
```json
{
  "provider": {
    "name": "azure",
    "base_url": "https://my-resource.openai.azure.com/openai",
    "options": {
      "deployment": "gpt-4o",
      "auth_type": "api_key",
      "token": "your-api-key",
      "api_version": "2025-01-01-preview"
    }
  }
}
```

### System-Assigned Managed Identity
```json
{
  "provider": {
    "name": "azure",
    "base_url": "https://my-resource.openai.azure.com/openai",
    "options": {
      "deployment": "gpt-4o",
      "auth_type": "managed_identity",
      "api_version": "2025-01-01-preview"
    }
  }
}
```

### User-Assigned Managed Identity
```json
{
  "provider": {
    "name": "azure",
    "base_url": "https://my-resource.openai.azure.com/openai",
    "options": {
      "deployment": "gpt-4o",
      "auth_type": "managed_identity",
      "api_version": "2025-01-01-preview",
      "resource": "https://cognitiveservices.azure.com/.default",
      "client_id": "your-managed-identity-client-id"
    }
  }
}
```

## Herald Impact

Once the new go-agents version is released:
- Remove the `newAgentFactory` workaround in `internal/infrastructure/infrastructure.go`
- The eager validation at `agent.New(&cfg.Agent)` will work because the provider accepts `managed_identity` without a static token
- Set the `HERALD_AGENT_TOKEN` environment variable to empty or remove it
- Optionally configure `resource` and `client_id` in the agent config if using user-assigned identity
