package logs

import (
	"context"
	"fmt"

	term2html "github.com/buildkite/terminal-to-html"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

const (
	STX = 0x02 // marks the beginning of logs for a phase
	ETX = 0x03 // marks the end of logs for a phase
)

type (
	// Chunk is a section of logs for a phase.
	Chunk struct {
		resource.TfeID `json:"id"` // Uniquely identifies the chunk.

		RunID  resource.TfeID     `json:"run_id"` // ID of run that generated the chunk
		Phase  internal.PhaseType `json:"phase"`  // Phase that generated the chunk
		Offset int                `json:"offset"` // Position within logs.
		Data   []byte             `json:"data"`   // The log data
	}

	PutChunkOptions struct {
		RunID  resource.TfeID     `schema:"run_id,required"`
		Phase  internal.PhaseType `schema:"phase,required"`
		Offset int                `schema:"offset,required"`
		Data   []byte
	}

	GetChunkOptions struct {
		RunID  resource.TfeID     `schema:"run_id"`
		Phase  internal.PhaseType `schema:"phase"`
		Limit  int                `schema:"limit"`  // size of the chunk to retrieve
		Offset int                `schema:"offset"` // position in overall data to seek from.
	}

	PutChunkService interface {
		PutChunk(ctx context.Context, opts PutChunkOptions) error
	}
)

func newChunk(opts PutChunkOptions) (Chunk, error) {
	if len(opts.Data) == 0 {
		return Chunk{}, fmt.Errorf("cowardly refusing to create empty log chunk")
	}
	chunk := Chunk{
		TfeID:  resource.NewTfeID(resource.ChunkKind),
		RunID:  opts.RunID,
		Phase:  opts.Phase,
		Offset: opts.Offset,
		Data:   opts.Data,
	}
	return chunk, nil
}

// Cut returns a new, smaller chunk.
func (c Chunk) Cut(opts GetChunkOptions) Chunk {
	if opts.Offset > c.NextOffset() {
		// offset is out of bounds - return an empty chunk with offset set to
		// the end of the chunk
		return Chunk{Offset: c.NextOffset()}
	}
	// ensure limit is not greater than the chunk itself.
	if (opts.Offset+opts.Limit) > c.NextOffset() || opts.Limit == 0 {
		opts.Limit = c.NextOffset() - opts.Offset
	}

	c.Data = c.Data[(opts.Offset - c.Offset):((opts.Offset - c.Offset) + opts.Limit)]
	c.Offset = opts.Offset

	return c
}

// NextOffset returns the offset for the next chunk
func (c Chunk) NextOffset() int {
	return c.Offset + len(c.Data)
}

func (c Chunk) IsStart() bool {
	return len(c.Data) > 0 && c.Data[0] == STX
}

func (c Chunk) IsEnd() bool {
	return len(c.Data) > 0 && c.Data[len(c.Data)-1] == ETX
}

func (c Chunk) ToHTML() string {
	// remove ASCII markers
	if c.IsStart() {
		c.Data = c.Data[1:]
	}
	if c.IsEnd() {
		c.Data = c.Data[:len(c.Data)-1]
	}

	// convert ANSI escape sequences to HTML
	html := term2html.Render(c.Data)

	return string(html)
}
