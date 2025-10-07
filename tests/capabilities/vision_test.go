package capabilities_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNewVisionCapability(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "images", Required: true, DefaultValue: nil},
		{Option: "detail", Required: false, DefaultValue: "auto"},
	}

	cap := capabilities.NewVisionCapability("openai-vision", options)

	if cap.Name() != "openai-vision" {
		t.Errorf("got name %q, want %q", cap.Name(), "openai-vision")
	}

	if cap.Protocol() != protocols.Vision {
		t.Errorf("got protocol %q, want %q", cap.Protocol(), protocols.Vision)
	}

	if !cap.SupportsStreaming() {
		t.Error("VisionCapability should support streaming")
	}
}

func TestVisionCapability_ProcessImages_SingleImage(t *testing.T) {
	cap := capabilities.NewVisionCapability("openai-vision", nil)

	messages := []protocols.Message{
		protocols.NewMessage("user", "What's in this image?"),
	}

	options := map[string]any{
		"images": []any{"https://example.com/image.jpg"},
	}

	result, err := cap.ProcessImages(messages, options)
	if err != nil {
		t.Fatalf("ProcessImages failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("got %d messages, want 1", len(result))
	}

	// Check that content was transformed to array format
	content, ok := result[0].Content.([]map[string]any)
	if !ok {
		t.Fatal("content is not array of maps")
	}

	if len(content) != 2 {
		t.Fatalf("got %d content items, want 2 (text + image)", len(content))
	}

	// Check text content
	if content[0]["type"] != "text" {
		t.Errorf("got content[0] type %v, want %q", content[0]["type"], "text")
	}

	// Check image content
	if content[1]["type"] != "image_url" {
		t.Errorf("got content[1] type %v, want %q", content[1]["type"], "image_url")
	}
}

func TestVisionCapability_ProcessImages_MultipleImages(t *testing.T) {
	cap := capabilities.NewVisionCapability("openai-vision", nil)

	messages := []protocols.Message{
		protocols.NewMessage("user", "Compare these images"),
	}

	options := map[string]any{
		"images": []any{
			"https://example.com/image1.jpg",
			"https://example.com/image2.jpg",
			"https://example.com/image3.jpg",
		},
	}

	result, err := cap.ProcessImages(messages, options)
	if err != nil {
		t.Fatalf("ProcessImages failed: %v", err)
	}

	content, ok := result[0].Content.([]map[string]any)
	if !ok {
		t.Fatal("content is not array of maps")
	}

	if len(content) != 4 {
		t.Fatalf("got %d content items, want 4 (text + 3 images)", len(content))
	}
}

func TestVisionCapability_ProcessImages_WithDetail(t *testing.T) {
	cap := capabilities.NewVisionCapability("openai-vision", nil)

	messages := []protocols.Message{
		protocols.NewMessage("user", "Analyze this"),
	}

	options := map[string]any{
		"images": []any{"https://example.com/image.jpg"},
		"detail": "high",
	}

	result, err := cap.ProcessImages(messages, options)
	if err != nil {
		t.Fatalf("ProcessImages failed: %v", err)
	}

	content, ok := result[0].Content.([]map[string]any)
	if !ok {
		t.Fatal("content is not array of maps")
	}

	imageContent := content[1]
	imageURL, ok := imageContent["image_url"].(map[string]any)
	if !ok {
		t.Fatal("image_url is not a map")
	}

	if imageURL["detail"] != "high" {
		t.Errorf("got detail %v, want %q", imageURL["detail"], "high")
	}

	// Verify detail was removed from options
	if _, exists := options["detail"]; exists {
		t.Error("detail should be removed from options after processing")
	}
}

func TestVisionCapability_ProcessImages_EmptyImages(t *testing.T) {
	cap := capabilities.NewVisionCapability("openai-vision", nil)

	messages := []protocols.Message{
		protocols.NewMessage("user", "What's in this image?"),
	}

	tests := []struct {
		name    string
		options map[string]any
	}{
		{
			name:    "missing images",
			options: map[string]any{},
		},
		{
			name:    "empty array",
			options: map[string]any{"images": []any{}},
		},
		{
			name:    "wrong type",
			options: map[string]any{"images": "not-an-array"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cap.ProcessImages(messages, tt.options)
			if err == nil {
				t.Error("expected error for invalid images, got nil")
			}
		})
	}
}

func TestVisionCapability_ProcessImages_EmptyMessages(t *testing.T) {
	cap := capabilities.NewVisionCapability("openai-vision", nil)

	messages := []protocols.Message{}

	options := map[string]any{
		"images": []any{"https://example.com/image.jpg"},
	}

	_, err := cap.ProcessImages(messages, options)
	if err == nil {
		t.Error("expected error for empty messages, got nil")
	}
}

func TestVisionCapability_ProcessImages_NonUserLastMessage(t *testing.T) {
	cap := capabilities.NewVisionCapability("openai-vision", nil)

	messages := []protocols.Message{
		protocols.NewMessage("user", "What's in this image?"),
		protocols.NewMessage("assistant", "I see a cat"),
	}

	options := map[string]any{
		"images": []any{"https://example.com/image.jpg"},
	}

	_, err := cap.ProcessImages(messages, options)
	if err == nil {
		t.Error("expected error for non-user last message, got nil")
	}
}

func TestVisionCapability_CreateRequest(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "images", Required: true, DefaultValue: nil},
		{Option: "temperature", Required: false, DefaultValue: 0.7},
	}

	cap := capabilities.NewVisionCapability("openai-vision", options)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Vision,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Describe this image"),
		},
		Options: map[string]any{
			"images":      []any{"https://example.com/image.jpg"},
			"temperature": 0.8,
		},
	}

	protocolReq, err := cap.CreateRequest(req, "gpt-4-vision")
	if err != nil {
		t.Fatalf("CreateRequest failed: %v", err)
	}

	if len(protocolReq.Messages) != 1 {
		t.Errorf("got %d messages, want 1", len(protocolReq.Messages))
	}

	if model, exists := protocolReq.Options["model"]; !exists {
		t.Error("model option missing")
	} else if model != "gpt-4-vision" {
		t.Errorf("got model %q, want %q", model, "gpt-4-vision")
	}

	// Verify images was removed from options
	if _, exists := protocolReq.Options["images"]; exists {
		t.Error("images should be removed from options after processing")
	}
}

func TestVisionCapability_CreateStreamingRequest(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "images", Required: true, DefaultValue: nil},
	}

	cap := capabilities.NewVisionCapability("openai-vision", options)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Vision,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Describe this image"),
		},
		Options: map[string]any{
			"images": []any{"https://example.com/image.jpg"},
		},
	}

	protocolReq, err := cap.CreateStreamingRequest(req, "gpt-4-vision")
	if err != nil {
		t.Fatalf("CreateStreamingRequest failed: %v", err)
	}

	if stream, exists := protocolReq.Options["stream"]; !exists {
		t.Error("stream option missing")
	} else if stream != true {
		t.Errorf("got stream %v, want true", stream)
	}
}

func TestVisionCapability_ParseResponse(t *testing.T) {
	cap := capabilities.NewVisionCapability("openai-vision", nil)

	responseData := []byte(`{
		"id": "chatcmpl-123",
		"model": "gpt-4-vision",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "The image shows a cat"
			},
			"finish_reason": "stop"
		}]
	}`)

	result, err := cap.ParseResponse(responseData)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	response, ok := result.(*protocols.ChatResponse)
	if !ok {
		t.Fatal("result is not a ChatResponse")
	}

	if content := response.Content(); content != "The image shows a cat" {
		t.Errorf("got content %q, want %q", content, "The image shows a cat")
	}
}
