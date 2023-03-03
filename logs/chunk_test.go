package logs

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestChunkCut(t *testing.T) {
	tests := []struct {
		name  string
		chunk otf.Chunk
		opts  otf.GetChunkOptions
		want  otf.Chunk
	}{
		{
			name: "cut nothing",
			chunk: otf.Chunk{
				Data: []byte("\x02hello world\x03"),
			},
			want: otf.Chunk{
				Data: []byte("\x02hello world\x03"),
			},
		},
		{
			name: "cut middle",
			chunk: otf.Chunk{
				Data: []byte("\x02hello world\x03"),
			},
			opts: otf.GetChunkOptions{Offset: 3, Limit: 4},
			want: otf.Chunk{
				Data:   []byte("llo "),
				Offset: 3,
			},
		},
		{
			name: "sanitize excessive limit",
			chunk: otf.Chunk{
				Data: []byte("\x02hello world\x03"),
			},
			opts: otf.GetChunkOptions{Offset: 3, Limit: 99},
			want: otf.Chunk{
				Data:   []byte("llo world\x03"),
				Offset: 3,
			},
		},
		{
			name: "handle excessive offset",
			chunk: otf.Chunk{
				Data: []byte("\x02hello world\x03"),
			},
			opts: otf.GetChunkOptions{Offset: 99},
			want: otf.Chunk{
				Offset: 13,
			},
		},
		{
			name: "cut chunk with non-zero offset",
			chunk: otf.Chunk{
				Data:   []byte(" world\x03"),
				Offset: 6,
			},
			opts: otf.GetChunkOptions{Offset: 8, Limit: 12},
			want: otf.Chunk{
				Data:   []byte("orld\x03"),
				Offset: 8,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.chunk.Cut(tt.opts)
			assert.Equal(t, tt.want, got)
		})
	}
}
