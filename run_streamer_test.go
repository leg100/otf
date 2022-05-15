package otf

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunStreamer(t *testing.T) {
	tests := []struct {
		name string
		run  Run
		want string
	}{
		{
			name: "plan with no changes",
			run: Run{
				Plan:  &Plan{ID: "plan-123"},
				Apply: &Apply{ID: "plan-123"},
			},
			want: "cat sat on the mat",
		},
		{
			name: "speculative plan",
			run: Run{
				Plan:  &Plan{ID: "plan-123", ResourceReport: &ResourceReport{ResourceAdditions: 1}},
				Apply: &Apply{ID: "plan-123"},
				ConfigurationVersion: &ConfigurationVersion{
					Speculative: true,
				},
			},
			want: "cat sat on the mat",
		},
		{
			name: "plan and apply",
			run: Run{
				Plan:  &Plan{ID: "plan-123", ResourceReport: &ResourceReport{ResourceAdditions: 1}},
				Apply: &Apply{ID: "apply-123"},
				ConfigurationVersion: &ConfigurationVersion{
					Speculative: false,
				},
			},
			want: "cat sat on the mat\nand then ate some food",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := testChunkStore{
				store: map[string]Chunk{
					"plan-123": Chunk{
						Start: true,
						End:   true,
						Data:  []byte("cat sat on the mat"),
					},
					"apply-123": Chunk{
						Start: true,
						End:   true,
						Data:  []byte("and then ate some food"),
					},
				},
			}

			streamer := NewRunStreamer(&tt.run, &store, &store, time.Millisecond)
			go streamer.Stream(context.Background())

			got, err := io.ReadAll(streamer)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(got))
		})
	}
}
