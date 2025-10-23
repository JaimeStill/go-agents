package classify

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func parseClassificationResponse(content string) (DocumentClassification, error) {
	var result DocumentClassification

	err1 := json.Unmarshal([]byte(content), &result)
	if err1 == nil {
		return result, nil
	}

	extracted := extractJSONFromMarkdown(content)
	err2 := json.Unmarshal([]byte(extracted), &result)
	if err2 == nil {
		return result, nil
	}

	return DocumentClassification{}, fmt.Errorf(
		"failed to parse classification response: %w",
		errors.Join(
			fmt.Errorf("direct parse: %w", err1),
			fmt.Errorf("markdown extraction: %w", err2),
		),
	)
}

func extractJSONFromMarkdown(content string) string {
	re := regexp.MustCompile(`(?s)` + "`" + `{3}(?:json)?\s*(.+?)\s*` + "`" + `{3}`)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return content
}
