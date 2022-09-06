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
	client := sse.NewClient(u.String())
	client.EncodingBase64 = true

	ch := make(chan otf.Event, 1)
	err = client.SubscribeRawWithContext(ctx, func(msg *sse.Event) {
		event := string(msg.Event)
		if strings.HasPrefix(event, "run_") {
			// bytes -> DTO
			dto := dto.Run{}
			err := jsonapi.UnmarshalPayload(bytes.NewReader(msg.Data), &dto)
			// handle error
			// DTO -> Domain
		}
	})
	return ch, err
}
