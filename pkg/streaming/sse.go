package streaming

import (
	"bufio"
	"context"
	"io"
	"strings"
)

// SSEMedia is the MIME type for Server-Sent Events streams.
const SSEMedia = "text/event-stream"

// SSEReader reads Server-Sent Events streams, parsing "data:" prefixed lines.
type SSEReader struct{}

// NewSSEReader returns a new SSEReader.
func NewSSEReader() *SSEReader {
	return &SSEReader{}
}

// ReadStream parses an SSE stream, emitting each data line as a StreamLine.
func (r *SSEReader) ReadStream(
	ctx context.Context,
	reader io.Reader,
) <-chan StreamLine {
	output := make(chan StreamLine)

	go func() {
		defer close(output)

		scanner := bufio.NewReader(reader)

		for {
			line, err := scanner.ReadString('\n')
			if err == io.EOF {
				return
			}
			if err != nil {
				select {
				case output <- StreamLine{Err: err}:
				case <-ctx.Done():
				}
				return
			}

			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			data, ok := strings.CutPrefix(line, "data: ")
			if !ok {
				continue
			}

			if data == "[DONE]" {
				select {
				case output <- StreamLine{Done: true}:
				case <-ctx.Done():
				}
				return
			}

			select {
			case output <- StreamLine{Data: []byte(data)}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}
