package otf

import "context"

type PersistedChunk interface {
	ID() string
	String() string
}

type LogsDB interface {
	// GetChunkByID fetches a specific chunk with the given ID.
	GetChunkByID(ctx context.Context, id int) (PersistedChunk, error)
}
