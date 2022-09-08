package http

import (
	"bytes"
	"context"

	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
	"github.com/r3labs/sse/v2"
)

func (c *client) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
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
				dto := dto.Run{}
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
			ch <- otf.Event{Type: otf.EventError, Payload: err.Error()}
		}
	}()
	return ch, nil
}
