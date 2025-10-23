package encoding_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/document"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/encoding"
)

func TestEncodeImageDataURI_PNG(t *testing.T) {
	imageData := []byte("fake-png-data")

	dataURI, err := encoding.EncodeImageDataURI(imageData, document.PNG)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify data URI structure for PNG
	if !strings.HasPrefix(dataURI, "data:image/png;base64,") {
		t.Errorf("expected PNG data URI prefix, got: %s", dataURI[:min(30, len(dataURI))])
	}

	// Verify base64 encoding is present
	if len(dataURI) <= len("data:image/png;base64,") {
		t.Error("data URI missing base64 content")
	}

	// Verify the base64 portion decodes correctly
	base64Part := strings.TrimPrefix(dataURI, "data:image/png;base64,")
	decoded, err := base64.StdEncoding.DecodeString(base64Part)
	if err != nil {
		t.Errorf("failed to decode base64 content: %v", err)
	}

	if string(decoded) != "fake-png-data" {
		t.Errorf("decoded content mismatch: got %q, want %q", string(decoded), "fake-png-data")
	}
}

func TestEncodeImageDataURI_JPEG(t *testing.T) {
	imageData := []byte("fake-jpeg-data")

	dataURI, err := encoding.EncodeImageDataURI(imageData, document.JPEG)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify data URI structure for JPEG
	if !strings.HasPrefix(dataURI, "data:image/jpeg;base64,") {
		t.Errorf("expected JPEG data URI prefix, got: %s", dataURI[:min(30, len(dataURI))])
	}

	// Verify base64 encoding is present
	if len(dataURI) <= len("data:image/jpeg;base64,") {
		t.Error("data URI missing base64 content")
	}
}

func TestEncodeImageDataURI_EmptyData(t *testing.T) {
	imageData := []byte{}

	_, err := encoding.EncodeImageDataURI(imageData, document.PNG)
	if err == nil {
		t.Fatal("expected error for empty image data")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' in error message, got: %v", err)
	}
}

func TestEncodeImageDataURI_UnsupportedFormat(t *testing.T) {
	imageData := []byte("test-data")
	invalidFormat := document.ImageFormat("bmp")

	_, err := encoding.EncodeImageDataURI(imageData, invalidFormat)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}

	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error message, got: %v", err)
	}
}

func TestEncodeImageDataURI_LargeData(t *testing.T) {
	// Test with larger data to ensure encoding works correctly
	imageData := make([]byte, 10000)
	for i := range imageData {
		imageData[i] = byte(i % 256)
	}

	dataURI, err := encoding.EncodeImageDataURI(imageData, document.PNG)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we can decode it back
	base64Part := strings.TrimPrefix(dataURI, "data:image/png;base64,")
	decoded, err := base64.StdEncoding.DecodeString(base64Part)
	if err != nil {
		t.Errorf("failed to decode large data: %v", err)
	}

	if len(decoded) != len(imageData) {
		t.Errorf("decoded length mismatch: got %d, want %d", len(decoded), len(imageData))
	}
}
