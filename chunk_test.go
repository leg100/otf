package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				Data:  []byte{'\x68', '\x65', '\x6c', '\x6c', '\x6f', '\x20', '\x77', '\x6f', '\x72', '\x6c', '\x6c', '\x64'},
				Start: true,
				End:   true,
			},
			want: Chunk{
				Data:  []byte{'\x68', '\x65', '\x6c', '\x6c', '\x6f', '\x20', '\x77', '\x6f', '\x72', '\x6c', '\x6c', '\x64'},
				Start: true,
				End:   true,
			},
		},
		{
			name: "get middle chunk",
			chunk: Chunk{
				Data:  []byte{'\x68', '\x65', '\x6c', '\x6c', '\x6f', '\x20', '\x77', '\x6f', '\x72', '\x6c', '\x6c', '\x64'},
				Start: true,
				End:   true,
			},
			opts: GetChunkOptions{Offset: 3, Limit: 4},
			want: Chunk{
				Data:  []byte{'\x6c', '\x6f', '\x20', '\x77'},
				Start: false,
				End:   false,
			},
		},
		{
			name: "get chunk with limit beyond size of data",
			chunk: Chunk{
				Data:  []byte{'\x68', '\x65', '\x6c', '\x6c', '\x6f', '\x20', '\x77', '\x6f', '\x72', '\x6c', '\x6c', '\x64'},
				Start: true,
				End:   true,
			},
			opts: GetChunkOptions{Offset: 3, Limit: 99},
			want: Chunk{
				Data:  []byte{'\x6c', '\x6f', '\x20', '\x77', '\x6f', '\x72', '\x6c', '\x6c', '\x64'},
				Start: true,
				End:   true,
			},
		},
		{
			name: "get chunk with offset beyond size of data",
			chunk: Chunk{
				Data:  []byte{'\x68', '\x65', '\x6c', '\x6c', '\x6f', '\x20', '\x77', '\x6f', '\x72', '\x6c', '\x6c', '\x64'},
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
