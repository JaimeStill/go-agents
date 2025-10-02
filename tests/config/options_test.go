package config_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
)

func TestExtractOption_ExistsWithCorrectType(t *testing.T) {
	options := map[string]any{
		"temperature": 0.7,
		"max_tokens":  4096,
		"model":       "gpt-4",
	}

	tests := []struct {
		name         string
		key          string
		defaultValue any
		expected     any
	}{
		{
			name:         "float64",
			key:          "temperature",
			defaultValue: 0.5,
			expected:     0.7,
		},
		{
			name:         "int",
			key:          "max_tokens",
			defaultValue: 1000,
			expected:     4096,
		},
		{
			name:         "string",
			key:          "model",
			defaultValue: "default",
			expected:     "gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case float64:
				result := config.ExtractOption(options, tt.key, tt.defaultValue.(float64))
				if result != expected {
					t.Errorf("got %v, want %v", result, expected)
				}
			case int:
				result := config.ExtractOption(options, tt.key, tt.defaultValue.(int))
				if result != expected {
					t.Errorf("got %v, want %v", result, expected)
				}
			case string:
				result := config.ExtractOption(options, tt.key, tt.defaultValue.(string))
				if result != expected {
					t.Errorf("got %v, want %v", result, expected)
				}
			}
		})
	}
}

func TestExtractOption_ExistsWithWrongType(t *testing.T) {
	options := map[string]any{
		"temperature": "0.7", // string instead of float64
	}

	result := config.ExtractOption(options, "temperature", 0.5)
	if result != 0.5 {
		t.Errorf("expected default value 0.5, got %v", result)
	}
}

func TestExtractOption_DoesNotExist(t *testing.T) {
	options := map[string]any{
		"temperature": 0.7,
	}

	result := config.ExtractOption(options, "nonexistent", 0.5)
	if result != 0.5 {
		t.Errorf("expected default value 0.5, got %v", result)
	}
}

func TestExtractOption_NilOptions(t *testing.T) {
	result := config.ExtractOption[float64](nil, "temperature", 0.5)
	if result != 0.5 {
		t.Errorf("expected default value 0.5, got %v", result)
	}
}

func TestMergeOptions(t *testing.T) {
	tests := []struct {
		name     string
		defaults map[string]any
		options  map[string]any
		expected map[string]any
	}{
		{
			name: "merge with overrides",
			defaults: map[string]any{
				"temperature": 0.5,
				"max_tokens":  1000,
			},
			options: map[string]any{
				"temperature": 0.7,
				"top_p":       0.95,
			},
			expected: map[string]any{
				"temperature": 0.7,
				"max_tokens":  1000,
				"top_p":       0.95,
			},
		},
		{
			name: "empty options",
			defaults: map[string]any{
				"temperature": 0.5,
			},
			options: map[string]any{},
			expected: map[string]any{
				"temperature": 0.5,
			},
		},
		{
			name:     "empty defaults",
			defaults: map[string]any{},
			options: map[string]any{
				"temperature": 0.7,
			},
			expected: map[string]any{
				"temperature": 0.7,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.MergeOptions(tt.defaults, tt.options)

			if len(result) != len(tt.expected) {
				t.Fatalf("result length %d does not match expected length %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				resultValue, exists := result[key]
				if !exists {
					t.Errorf("key %s missing from result", key)
					continue
				}
				if resultValue != expectedValue {
					t.Errorf("key %s: got %v, want %v", key, resultValue, expectedValue)
				}
			}
		})
	}
}

func TestFilterSupportedOptions(t *testing.T) {
	tests := []struct {
		name      string
		options   map[string]any
		supported []string
		expected  map[string]any
	}{
		{
			name: "filters unsupported options",
			options: map[string]any{
				"temperature":        0.7,
				"max_tokens":         4096,
				"unsupported_option": "value",
			},
			supported: []string{"temperature", "max_tokens"},
			expected: map[string]any{
				"temperature": 0.7,
				"max_tokens":  4096,
			},
		},
		{
			name: "all options supported",
			options: map[string]any{
				"temperature": 0.7,
				"max_tokens":  4096,
			},
			supported: []string{"temperature", "max_tokens"},
			expected: map[string]any{
				"temperature": 0.7,
				"max_tokens":  4096,
			},
		},
		{
			name: "no options supported",
			options: map[string]any{
				"temperature": 0.7,
				"max_tokens":  4096,
			},
			supported: []string{},
			expected:  map[string]any{},
		},
		{
			name:      "empty options",
			options:   map[string]any{},
			supported: []string{"temperature"},
			expected:  map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.FilterSupportedOptions(tt.options, tt.supported)

			if len(result) != len(tt.expected) {
				t.Fatalf("result length %d does not match expected length %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				resultValue, exists := result[key]
				if !exists {
					t.Errorf("key %s missing from result", key)
					continue
				}
				if resultValue != expectedValue {
					t.Errorf("key %s: got %v, want %v", key, resultValue, expectedValue)
				}
			}
		})
	}
}

func TestValidateRequiredOptions(t *testing.T) {
	tests := []struct {
		name        string
		options     map[string]any
		required    []string
		expectError bool
	}{
		{
			name: "all required present",
			options: map[string]any{
				"temperature": 0.7,
				"max_tokens":  4096,
			},
			required:    []string{"temperature", "max_tokens"},
			expectError: false,
		},
		{
			name: "missing required option",
			options: map[string]any{
				"temperature": 0.7,
			},
			required:    []string{"temperature", "max_tokens"},
			expectError: true,
		},
		{
			name: "no required options",
			options: map[string]any{
				"temperature": 0.7,
			},
			required:    []string{},
			expectError: false,
		},
		{
			name:        "empty options with required",
			options:     map[string]any{},
			required:    []string{"temperature"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateRequiredOptions(tt.options, tt.required)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
