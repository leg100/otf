package http

import (
	"bytes"
	"context"
	"strings"

	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
	"github.com/r3labs/sse/v2"
)

func (c *client) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	u, err := c.baseURL.Parse("/watch")
	if err != nil {
		return nil, err
	}
	sseClient := newSSEClient(u.String(), c.insecure)
	ch := make(chan otf.Event, 1)

	err = sseClient.SubscribeRawWithContext(ctx, func(msg *sse.Event) {
		event := string(msg.Event)
		// TODO: impl support for objects other than runs
		if strings.HasPrefix(event, "run_") {
			// bytes -> DTO
			dto := dto.Run{}
			if err := jsonapi.UnmarshalPayload(bytes.NewReader(msg.Data), &dto); err != nil {
				ch <- otf.Event{Type: otf.EventError, Payload: err.Error()}
				return
			}
			// DTO -> Domain
			run := otf.UnmarshalRunJSONAPI(&dto)

			ch <- otf.Event{Type: otf.EventType(event), Payload: run}
		}
	})
	return ch, err
}
