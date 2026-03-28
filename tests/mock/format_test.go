package mock_test

import (
	"errors"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

func TestNewMockFormat(t *testing.T) {
	f := mock.NewMockFormat()

	if f == nil {
		t.Fatal("NewMockFormat returned nil")
	}
}

func TestMockFormat_Name(t *testing.T) {
	f := mock.NewMockFormat()

	if f.Name() != "mock-format" {
		t.Errorf("got name %q, want %q", f.Name(), "mock-format")
	}
}

func TestMockFormat_Name_Custom(t *testing.T) {
	f := mock.NewMockFormat(
		mock.WithFormatName("custom-format"),
	)

	if f.Name() != "custom-format" {
		t.Errorf("got name %q, want %q", f.Name(), "custom-format")
	}
}

func TestMockFormat_Marshal(t *testing.T) {
	expectedBody := []byte(`{"model":"test-model"}`)

	f := mock.NewMockFormat(
		mock.WithFormatMarshalResponse(expectedBody, nil),
	)

	body, err := f.Marshal(protocol.Chat, nil)

	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(body) != string(expectedBody) {
		t.Errorf("got body %q, want %q", string(body), string(expectedBody))
	}
}

func TestMockFormat_Marshal_Error(t *testing.T) {
	expectedErr := errors.New("marshal failed")

	f := mock.NewMockFormat(
		mock.WithFormatMarshalResponse(nil, expectedErr),
	)

	_, err := f.Marshal(protocol.Chat, nil)

	if err == nil {
		t.Fatal("expected error from Marshal, got nil")
	}

	if err != expectedErr {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestMockFormat_Marshal_Default(t *testing.T) {
	f := mock.NewMockFormat()

	body, err := f.Marshal(protocol.Chat, nil)

	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(body) != `{}` {
		t.Errorf("got body %q, want %q", string(body), `{}`)
	}
}

func TestMockFormat_Parse(t *testing.T) {
	expectedResp := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "test"},
		},
	}

	f := mock.NewMockFormat(
		mock.WithFormatParseResponse(expectedResp, nil),
	)

	result, err := f.Parse(protocol.Chat, nil)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result != expectedResp {
		t.Error("returned different response than configured")
	}
}

func TestMockFormat_Parse_Error(t *testing.T) {
	expectedErr := errors.New("parse failed")

	f := mock.NewMockFormat(
		mock.WithFormatParseResponse(nil, expectedErr),
	)

	_, err := f.Parse(protocol.Chat, nil)

	if err == nil {
		t.Fatal("expected error from Parse, got nil")
	}

	if err != expectedErr {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestMockFormat_ParseStreamChunk(t *testing.T) {
	expectedChunk := &response.StreamingResponse{
		Content: []response.ContentBlock{
			response.TextBlock{Text: "test"},
		},
	}

	f := mock.NewMockFormat(
		mock.WithFormatStreamChunk(expectedChunk, nil),
	)

	chunk, err := f.ParseStreamChunk(protocol.Chat, nil)

	if err != nil {
		t.Fatalf("ParseStreamChunk failed: %v", err)
	}

	if chunk != expectedChunk {
		t.Error("returned different chunk than configured")
	}
}

func TestMockFormat_ParseStreamChunk_Error(t *testing.T) {
	expectedErr := errors.New("stream chunk failed")

	f := mock.NewMockFormat(
		mock.WithFormatStreamChunk(nil, expectedErr),
	)

	_, err := f.ParseStreamChunk(protocol.Chat, nil)

	if err == nil {
		t.Fatal("expected error from ParseStreamChunk, got nil")
	}

	if err != expectedErr {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}
