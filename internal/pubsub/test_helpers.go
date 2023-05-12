package pubsub

import (
	"context"

	"github.com/leg100/otf/internal"
)

type (
	fakePool struct {
		pool
	}

	fakeUnmarshaler struct {
		resource any
	}
)

func (f *fakeUnmarshaler) UnmarshalEvent(ctx context.Context, payload []byte, op internal.EventType) (any, error) {
	return f.resource, nil
}
