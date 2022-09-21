package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkCut(t *testing.T) {
	tests := []struct {
		name  string
		chunk Chunk
		opts  GetChunkOptions
		want  Chunk
	}{
		{
			name: "cut nothing",
			chunk: Chunk{
				Data: []byte("\x02hello world\x03"),
			},
			want: Chunk{
				Data: []byte("\x02hello world\x03"),
			},
		},
		{
			name: "cut middle",
			chunk: Chunk{
				Data: []byte("\x02hello world\x03"),
			},
			opts: GetChunkOptions{Offset: 3, Limit: 4},
			want: Chunk{
				Data:   []byte("llo "),
				Offset: 3,
			},
		},
		{
			name: "sanitize excessive limit",
			chunk: Chunk{
				Data: []byte("\x02hello world\x03"),
			},
			opts: GetChunkOptions{Offset: 3, Limit: 99},
			want: Chunk{
				Data:   []byte("llo world\x03"),
				Offset: 3,
			},
		},
		{
			name: "handle excessive offset",
			chunk: Chunk{
				Data: []byte("\x02hello world\x03"),
			},
			opts: GetChunkOptions{Offset: 99},
			want: Chunk{
				Offset: 13,
			},
		},
		{
			name: "cut chunk with non-zero offset",
			chunk: Chunk{
				Data:   []byte(" world\x03"),
				Offset: 6,
			},
			opts: GetChunkOptions{Offset: 8, Limit: 12},
			want: Chunk{
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
