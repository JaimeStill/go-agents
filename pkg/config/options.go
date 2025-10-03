package config

import (
	"fmt"
	"maps"
	"slices"
)

// ExtractOption retrieves a typed value from an options map.
// If the key exists and the value is of type T, it returns the value.
// Otherwise, it returns the provided default value.
func ExtractOption[T any](options map[string]any, key string, defaultValue T) T {
	if value, exists := options[key]; exists {
		if typedValue, ok := value.(T); ok {
			return typedValue
		}
	}
	return defaultValue
}

// MergeOptions combines two option maps, with values from options taking precedence
// over defaults. The defaults map is modified in place and returned.
func MergeOptions(defaults map[string]any, options map[string]any) map[string]any {
	maps.Copy(defaults, options)

	return defaults
}

// FilterSupportedOptions returns a new map containing only the options that are
// present in the supported list. Unsupported options are silently ignored.
func FilterSupportedOptions(options map[string]any, supported []string) map[string]any {
	filtered := make(map[string]any)

	for key, value := range options {
		if slices.Contains(supported, key) {
			filtered[key] = value
		}
	}

	return filtered
}

// ValidateRequiredOptions checks that all required options are present in the
// options map. Returns an error if any required option is missing.
func ValidateRequiredOptions(options map[string]any, required []string) error {
	for _, key := range required {
		if _, exists := options[key]; !exists {
			return fmt.Errorf("required option '%s' is missing", key)
		}
	}
	return nil
}
