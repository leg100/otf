package watch

import (
	"bytes"
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/r3labs/sse/v2"
)

func (c *Client) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	// TODO: why buffered chan of size 1?
	ch := make(chan otf.Event, 1)
	sseClient, err := c.newSSEClient("watch", ch)
	if err != nil {
		return nil, err
	}

	go func() {
		err = sseClient.SubscribeRawWithContext(ctx, func(msg *sse.Event) {
			// TODO: impl support for objects other than runs
			if bytes.HasPrefix(msg.Event, []byte("run_")) {
				// bytes -> DTO
				dto := jsonapi.Run{}
				if err := jsonapi.UnmarshalPayload(bytes.NewReader(msg.Data), &dto); err != nil {
					ch <- otf.Event{Type: otf.EventError, Payload: err.Error()}
					return
				}
				// DTO -> Domain
				run := otf.UnmarshalRunJSONAPI(&dto)

				ch <- otf.Event{Type: otf.EventType(msg.Event), Payload: run}
			}
		})
		if err != nil {
			ch <- otf.Event{Type: otf.EventError, Payload: err}
		}
		close(ch)
	}()
	return ch, nil
}
