package format_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/format"
)

func TestCreate_OpenAI(t *testing.T) {
	f, err := format.Create("openai")

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if f == nil {
		t.Fatal("Create returned nil format")
	}

	if f.Name() != "openai" {
		t.Errorf("got name %q, want %q", f.Name(), "openai")
	}
}

func TestCreate_UnknownFormat(t *testing.T) {
	_, err := format.Create("unknown-format")

	if err == nil {
		t.Error("expected error for unknown format, got nil")
	}
}

func TestCreate_Converse(t *testing.T) {
	f, err := format.Create("converse")

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if f == nil {
		t.Fatal("Create returned nil format")
	}

	if f.Name() != "converse" {
		t.Errorf("got name %q, want %q", f.Name(), "converse")
	}
}

func TestListFormats(t *testing.T) {
	names := format.ListFormats()

	if len(names) == 0 {
		t.Error("ListFormats returned empty list")
	}

	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}

	if !found["openai"] {
		t.Error("openai format not registered")
	}

	if !found["converse"] {
		t.Error("converse format not registered")
	}
}
