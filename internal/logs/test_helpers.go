package logs

import (
	"context"

	"github.com/leg100/otf/internal/pubsub"
)

type fakeSubService struct {
	stream chan pubsub.Event[Chunk]

	pubsub.SubscriptionService[Chunk]
}

func (f *fakeSubService) Subscribe(ctx context.Context) (<-chan pubsub.Event[Chunk], func()) {
	go func() {
		<-ctx.Done()
		close(f.stream)
	}()
	return f.stream, nil
}
