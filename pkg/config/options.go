package config

import (
	"fmt"
	"maps"
	"slices"
)

func ExtractOption[T any](options map[string]any, key string, defaultValue T) T {
	if value, exists := options[key]; exists {
		if typedValue, ok := value.(T); ok {
			return typedValue
		}
	}
	return defaultValue
}

func MergeOptions(defaults map[string]any, options map[string]any) map[string]any {
	maps.Copy(defaults, options)

	return defaults
}

func FilterSupportedOptions(options map[string]any, supported []string) map[string]any {
	filtered := make(map[string]any)

	for key, value := range options {
		if slices.Contains(supported, key) {
			filtered[key] = value
		}
	}

	return filtered
}

func ValidateRequiredOptions(options map[string]any, required []string) error {
	for _, key := range required {
		if _, exists := options[key]; !exists {
			return fmt.Errorf("required option '%s' is missing", key)
		}
	}
	return nil
}
