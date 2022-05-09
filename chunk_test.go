package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkMarshal(t *testing.T) {
	tests := []struct {
		name  string
		chunk Chunk
		want  string
	}{
		{
			name: "both start and end markers",
			chunk: Chunk{
				Data:  []byte("hello world"),
				Start: true,
				End:   true,
			},
			want: "\x02hello world\x03",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.chunk.Marshal()))
		})
	}
}

func TestChunkUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Chunk
	}{
		{
			name:  "start marker",
			input: "\x02hello",
			want: Chunk{
				Data:  []byte("hello"),
				Start: true,
				End:   false,
			},
		},
		{
			name:  "end marker",
			input: " world\x03",
			want: Chunk{
				Data:  []byte(" world"),
				Start: false,
				End:   true,
			},
		},
		{
			name:  "start and end marker",
			input: "\x02hello world\x03",
			want: Chunk{
				Data:  []byte("hello world"),
				Start: true,
				End:   true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, UnmarshalChunk([]byte(tt.input)))
		})
	}
}

func TestChunkCut(t *testing.T) {
	tests := []struct {
		name    string
		chunk   Chunk
		opts    GetChunkOptions
		want    Chunk
		wantErr bool
	}{
		{
			name: "get all data",
			chunk: Chunk{
				Data:  []byte("hello world"),
				Start: true,
				End:   true,
			},
			want: Chunk{
				Data:  []byte("hello world"),
				Start: true,
				End:   true,
			},
		},
		{
			name: "get middle chunk",
			chunk: Chunk{
				Data:  []byte("hello world"),
				Start: true,
				End:   true,
			},
			opts: GetChunkOptions{Offset: 3, Limit: 4},
			want: Chunk{
				Data:  []byte("llo "),
				Start: false,
				End:   false,
			},
		},
		{
			name: "get chunk with limit beyond size of data",
			chunk: Chunk{
				Data:  []byte("hello world"),
				Start: true,
				End:   true,
			},
			opts: GetChunkOptions{Offset: 3, Limit: 99},
			want: Chunk{
				Data:  []byte("llo world"),
				Start: false,
				End:   true,
			},
		},
		{
			name: "get chunk with offset beyond size of data",
			chunk: Chunk{
				Data:  []byte("hello world"),
				Start: true,
				End:   true,
			},
			opts:    GetChunkOptions{Offset: 99, Limit: 4},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.chunk.Cut(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
