package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
)

// EventStreamMedia is the MIME type for AWS event stream encoding.
const EventStreamMedia = "application/vnd.amazon.eventstream"

// EventStreamReader reads AWS event stream encoded responses.
type EventStreamReader struct{}

// NewEventStreamReader returns a new EventStreamReader.
func NewEventStreamReader() *EventStreamReader {
	return &EventStreamReader{}
}

// ReadStream decodes an AWS event stream, emitting each event as a StreamLine.
func (r *EventStreamReader) ReadStream(
	ctx context.Context,
	reader io.Reader,
) <-chan StreamLine {
	output := make(chan StreamLine)

	go func() {
		defer close(output)

		decoder := eventstream.NewDecoder()

		for {
			msg, err := decoder.Decode(reader, nil)
			if err == io.EOF {
				return
			}
			if err != nil {
				select {
				case output <- StreamLine{
					Err: fmt.Errorf("decode event stream frame: %w", err),
				}:
				case <-ctx.Done():
				}
				return
			}

			var messageType, eventType string
			for _, h := range msg.Headers {
				switch h.Name {
				case ":message-type":
					messageType = h.Value.String()
				case ":event-type":
					eventType = h.Value.String()
				}
			}

			if messageType == "exception" {
				select {
				case output <- StreamLine{
					Err: fmt.Errorf("event stream exception: %s: %s", eventType, string(msg.Payload)),
				}:
				case <-ctx.Done():
				}
				return
			}

			envelope, err := json.Marshal(map[string]json.RawMessage{
				eventType: msg.Payload,
			})
			if err != nil {
				continue
			}

			select {
			case output <- StreamLine{Data: envelope}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}
