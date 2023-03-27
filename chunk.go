package otf

import (
	"context"
	"html/template"

	term2html "github.com/buildkite/terminal-to-html"
)

const (
	STX = 0x02 // marks the beginning of logs for a phase
	ETX = 0x03 // marks the end of logs for a phase
)

type (
	// Chunk is a section of logs for a phase.
	Chunk struct {
		ID     string    // Uniquely identifies the chunk.
		RunID  string    // ID of run that generated the chunk
		Phase  PhaseType // Phase that generated the chunk
		Offset int       // Position within logs.
		Data   []byte    // The log data
	}

	PutChunkOptions struct {
		RunID  string    `schema:"run_id,required"`
		Phase  PhaseType `schema:"phase,required"`
		Offset int       `schema:"offset,required"`
		Data   []byte
	}

	GetChunkOptions struct {
		RunID  string    `schema:"run_id"`
		Phase  PhaseType `schema:"phase"`
		Limit  int       `schema:"limit"`  // size of the chunk to retrieve
		Offset int       `schema:"offset"` // position in overall data to seek from.
	}

	PutChunkService interface {
		PutChunk(ctx context.Context, opts PutChunkOptions) error
	}
)

// Cut returns a new, smaller chunk.
func (c Chunk) Cut(opts GetChunkOptions) Chunk {
	if opts.Offset > c.NextOffset() {
		// offset is out of bounds - return an empty chunk with offset set to
		// the end of the chunk
		return Chunk{Offset: c.NextOffset()}
	}
	// sanitize limit - 0 means limitless.
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

func (c Chunk) ToHTML() template.HTML {
	// remove ASCII markers
	if c.IsStart() {
		c.Data = c.Data[1:]
	}
	if c.IsEnd() {
		c.Data = c.Data[:len(c.Data)-1]
	}

	// convert ANSI escape sequences to HTML
	html := term2html.Render(c.Data)

	return template.HTML(string(html))
}
