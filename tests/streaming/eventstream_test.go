package streaming_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"

	"github.com/JaimeStill/go-agents/pkg/streaming"
)

// encodeEvent encodes a single event stream message into binary format.
func encodeEvent(messageType, eventType string, payload []byte) []byte {
	msg := eventstream.Message{
		Headers: eventstream.Headers{
			{
				Name:  ":message-type",
				Value: eventstream.StringValue(messageType),
			},
			{
				Name:  ":event-type",
				Value: eventstream.StringValue(eventType),
			},
		},
		Payload: payload,
	}

	var buf bytes.Buffer
	encoder := eventstream.NewEncoder()
	encoder.Encode(&buf, msg)
	return buf.Bytes()
}

func TestEventStreamReader_ContentBlockDelta(t *testing.T) {
	payload := []byte(`{"contentBlockIndex":0,"delta":{"text":"Hello"}}`)
	data := encodeEvent("event", "contentBlockDelta", payload)

	reader := streaming.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), bytes.NewReader(data))

	line := <-lines
	if line.Err != nil {
		t.Fatalf("unexpected error: %v", line.Err)
	}

	// The event stream reader wraps the payload in a JSON envelope keyed by event type
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(line.Data, &envelope); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	raw, ok := envelope["contentBlockDelta"]
	if !ok {
		t.Fatal("expected contentBlockDelta key in envelope")
	}

	var delta struct {
		ContentBlockIndex int `json:"contentBlockIndex"`
		Delta             struct {
			Text string `json:"text"`
		} `json:"delta"`
	}
	if err := json.Unmarshal(raw, &delta); err != nil {
		t.Fatalf("failed to unmarshal delta: %v", err)
	}

	if delta.Delta.Text != "Hello" {
		t.Errorf("got text %q, want %q", delta.Delta.Text, "Hello")
	}
}

func TestEventStreamReader_MessageStop(t *testing.T) {
	payload := []byte(`{"stopReason":"end_turn"}`)
	data := encodeEvent("event", "messageStop", payload)

	reader := streaming.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), bytes.NewReader(data))

	line := <-lines
	if line.Err != nil {
		t.Fatalf("unexpected error: %v", line.Err)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(line.Data, &envelope); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	raw, ok := envelope["messageStop"]
	if !ok {
		t.Fatal("expected messageStop key in envelope")
	}

	var stop struct {
		StopReason string `json:"stopReason"`
	}
	if err := json.Unmarshal(raw, &stop); err != nil {
		t.Fatalf("failed to unmarshal stop: %v", err)
	}

	if stop.StopReason != "end_turn" {
		t.Errorf("got stop reason %q, want %q", stop.StopReason, "end_turn")
	}
}

func TestEventStreamReader_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(encodeEvent("event", "contentBlockDelta", []byte(`{"delta":{"text":"Hello"}}`)))
	buf.Write(encodeEvent("event", "contentBlockDelta", []byte(`{"delta":{"text":" world"}}`)))
	buf.Write(encodeEvent("event", "messageStop", []byte(`{"stopReason":"end_turn"}`)))

	reader := streaming.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), &buf)

	count := 0
	for line := range lines {
		if line.Err != nil {
			t.Fatalf("unexpected error on line %d: %v", count, line.Err)
		}
		count++
	}

	if count != 3 {
		t.Errorf("got %d lines, want 3", count)
	}
}

func TestEventStreamReader_Exception(t *testing.T) {
	data := encodeEvent("exception", "validationException", []byte(`{"message":"Invalid model ID"}`))

	reader := streaming.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), bytes.NewReader(data))

	line := <-lines
	if line.Err == nil {
		t.Fatal("expected error for exception event, got nil")
	}

	// Channel should close after exception
	_, open := <-lines
	if open {
		t.Error("expected channel to be closed after exception")
	}
}

func TestEventStreamReader_EOF(t *testing.T) {
	// Empty reader — immediate EOF
	reader := streaming.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), bytes.NewReader(nil))

	_, open := <-lines
	if open {
		t.Error("expected channel to be closed on EOF")
	}
}
