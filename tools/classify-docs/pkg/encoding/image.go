package encoding

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/document"
)

func EncodeImageDataURI(data []byte, format document.ImageFormat) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("image data is empty")
	}

	mimeType, err := format.MimeType()
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("data:")
	builder.WriteString(mimeType)
	builder.WriteString(";base64,")
	builder.WriteString(base64.StdEncoding.EncodeToString(data))

	return builder.String(), nil
}
