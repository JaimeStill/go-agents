package identities_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/identities"
)

func TestNewAWSCredentialSource_Static(t *testing.T) {
	options := map[string]any{
		"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
		"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	source, err := identities.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		identities.AWSAuthStatic,
		options,
	)

	if err != nil {
		t.Fatalf("NewAWSCredentialSource failed: %v", err)
	}

	if source == nil {
		t.Fatal("NewAWSCredentialSource returned nil")
	}
}

func TestNewAWSCredentialSource_Static_WithSessionToken(t *testing.T) {
	options := map[string]any{
		"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
		"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"session_token":     "FwoGZXIvYXdzEBYaDH...",
	}

	source, err := identities.NewAWSCredentialSource(
		context.Background(),
		"us-west-2",
		identities.AWSAuthStatic,
		options,
	)

	if err != nil {
		t.Fatalf("NewAWSCredentialSource failed: %v", err)
	}

	if source == nil {
		t.Fatal("NewAWSCredentialSource returned nil")
	}
}

func TestNewAWSCredentialSource_Static_MissingAccessKey(t *testing.T) {
	options := map[string]any{
		"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	_, err := identities.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		identities.AWSAuthStatic,
		options,
	)

	if err == nil {
		t.Error("expected error for missing access_key_id, got nil")
	}
}

func TestNewAWSCredentialSource_Static_MissingSecretKey(t *testing.T) {
	options := map[string]any{
		"access_key_id": "AKIAIOSFODNN7EXAMPLE",
	}

	_, err := identities.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		identities.AWSAuthStatic,
		options,
	)

	if err == nil {
		t.Error("expected error for missing secret_access_key, got nil")
	}
}

func TestNewAWSCredentialSource_Profile_MissingProfile(t *testing.T) {
	_, err := identities.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		identities.AWSAuthProfile,
		map[string]any{},
	)

	if err == nil {
		t.Error("expected error for missing profile, got nil")
	}
}

func TestNewAWSCredentialSource_UnsupportedAuthType(t *testing.T) {
	_, err := identities.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		identities.AWSAuthType("unsupported"),
		map[string]any{},
	)

	if err == nil {
		t.Error("expected error for unsupported auth type, got nil")
	}
}

func TestSignRequest_AddsAuthorizationHeader(t *testing.T) {
	options := map[string]any{
		"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
		"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	source, err := identities.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		identities.AWSAuthStatic,
		options,
	)
	if err != nil {
		t.Fatalf("NewAWSCredentialSource failed: %v", err)
	}

	req, _ := http.NewRequest("POST", "https://bedrock-runtime.us-east-1.amazonaws.com/model/test/converse", strings.NewReader(`{"test":"data"}`))
	req.Header.Set("Content-Type", "application/json")

	err = source.SignRequest(context.Background(), req, "bedrock")
	if err != nil {
		t.Fatalf("SignRequest failed: %v", err)
	}

	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Error("Authorization header is empty after signing")
	}

	if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256") {
		t.Errorf("Authorization header doesn't start with SigV4 prefix, got %q", auth)
	}

	// Verify the credential scope includes the service name
	if !strings.Contains(auth, "bedrock") {
		t.Error("Authorization header doesn't contain service name 'bedrock'")
	}

	// Verify X-Amz-Date is set
	if req.Header.Get("X-Amz-Date") == "" {
		t.Error("X-Amz-Date header is empty after signing")
	}
}

func TestSignRequest_BodyRemainsReadable(t *testing.T) {
	options := map[string]any{
		"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
		"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	source, err := identities.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		identities.AWSAuthStatic,
		options,
	)
	if err != nil {
		t.Fatalf("NewAWSCredentialSource failed: %v", err)
	}

	bodyContent := `{"modelId":"anthropic.claude-sonnet-4-6-v1"}`
	req, _ := http.NewRequest("POST", "https://example.com/test", strings.NewReader(bodyContent))

	err = source.SignRequest(context.Background(), req, "bedrock")
	if err != nil {
		t.Fatalf("SignRequest failed: %v", err)
	}

	// Body should still be readable after signing
	buf := make([]byte, len(bodyContent))
	n, _ := req.Body.Read(buf)

	if string(buf[:n]) != bodyContent {
		t.Errorf("body after signing: got %q, want %q", string(buf[:n]), bodyContent)
	}
}
