package run

import (
	"context"

	"github.com/leg100/otf/internal/pubsub"
)

type tailerClient interface {
	GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error)
}

type tailer struct {
	client tailerClient
	broker pubsub.SubscriptionService[Chunk]
}

func (t *tailer) Tail(ctx context.Context, opts TailOptions) (<-chan Chunk, error) {
	// Subscribe first and only then retrieve from DB, guaranteeing that we
	// won't miss any updates
	sub, _ := t.broker.Subscribe(ctx)

	chunk, err := t.client.GetChunk(ctx, GetChunkOptions{
		RunID:  opts.RunID,
		Phase:  opts.Phase,
		Offset: opts.Offset,
	})
	if err != nil {
		return nil, err
	}
	opts.Offset += len(chunk.Data)

	// relay is the chan returned to the caller on which chunks are relayed to.
	relay := make(chan Chunk)
	go func() {
		// send existing chunk
		if len(chunk.Data) > 0 {
			relay <- chunk
		}

		// relay chunks from subscription
		for ev := range sub {
			chunk := ev.Payload
			if opts.RunID != chunk.RunID || opts.Phase != chunk.Phase {
				// skip logs for different run/phase
				continue
			}
			if chunk.Offset < opts.Offset {
				// chunk has overlapping offset
				if chunk.Offset+len(chunk.Data) <= opts.Offset {
					// skip entirely overlapping chunk
					continue
				}
				// remove overlapping portion of chunk
				chunk = chunk.Cut(GetChunkOptions{Offset: opts.Offset})
			}
			if len(chunk.Data) == 0 {
				// don't send empty chunks
				continue
			}
			relay <- chunk
			if chunk.IsEnd() {
				break
			}
		}
		close(relay)
	}()
	return relay, nil
}
