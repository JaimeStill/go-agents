package streaming_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/streaming"
)

func TestSSEReader_TextData(t *testing.T) {
	input := "data: {\"content\":\"hello\"}\n\n"
	reader := streaming.NewSSEReader()

	lines := reader.ReadStream(context.Background(), strings.NewReader(input))

	line := <-lines
	if line.Err != nil {
		t.Fatalf("unexpected error: %v", line.Err)
	}
	if line.Done {
		t.Fatal("unexpected done signal")
	}
	if string(line.Data) != `{"content":"hello"}` {
		t.Errorf("got data %q, want %q", string(line.Data), `{"content":"hello"}`)
	}
}

func TestSSEReader_MultipleLines(t *testing.T) {
	input := "data: first\ndata: second\n"
	reader := streaming.NewSSEReader()

	lines := reader.ReadStream(context.Background(), strings.NewReader(input))

	first := <-lines
	if string(first.Data) != "first" {
		t.Errorf("got first data %q, want %q", string(first.Data), "first")
	}

	second := <-lines
	if string(second.Data) != "second" {
		t.Errorf("got second data %q, want %q", string(second.Data), "second")
	}
}

func TestSSEReader_DoneSignal(t *testing.T) {
	input := "data: hello\ndata: [DONE]\n"
	reader := streaming.NewSSEReader()

	lines := reader.ReadStream(context.Background(), strings.NewReader(input))

	first := <-lines
	if string(first.Data) != "hello" {
		t.Errorf("got data %q, want %q", string(first.Data), "hello")
	}

	done := <-lines
	if !done.Done {
		t.Error("expected done signal")
	}

	// Channel should be closed after done
	_, open := <-lines
	if open {
		t.Error("expected channel to be closed after done")
	}
}

func TestSSEReader_SkipsEmptyLines(t *testing.T) {
	input := "\n\ndata: hello\n\n"
	reader := streaming.NewSSEReader()

	lines := reader.ReadStream(context.Background(), strings.NewReader(input))

	line := <-lines
	if string(line.Data) != "hello" {
		t.Errorf("got data %q, want %q", string(line.Data), "hello")
	}
}

func TestSSEReader_SkipsNonDataLines(t *testing.T) {
	input := "event: message\nid: 123\ndata: hello\n"
	reader := streaming.NewSSEReader()

	lines := reader.ReadStream(context.Background(), strings.NewReader(input))

	line := <-lines
	if string(line.Data) != "hello" {
		t.Errorf("got data %q, want %q", string(line.Data), "hello")
	}
}

func TestSSEReader_ReadError(t *testing.T) {
	reader := streaming.NewSSEReader()

	lines := reader.ReadStream(context.Background(), &errorReader{err: io.ErrUnexpectedEOF})

	line := <-lines
	if line.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSSEReader_EOF(t *testing.T) {
	input := "data: hello\n"
	reader := streaming.NewSSEReader()

	lines := reader.ReadStream(context.Background(), strings.NewReader(input))

	line := <-lines
	if string(line.Data) != "hello" {
		t.Errorf("got data %q, want %q", string(line.Data), "hello")
	}

	// Channel should close on EOF
	_, open := <-lines
	if open {
		t.Error("expected channel to be closed on EOF")
	}
}

// errorReader is a test helper that always returns an error.
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, r.err
}
