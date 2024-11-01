package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		runner      *RunnerMeta
		want        RunnerStatus
		wantDeleted bool
	}{
		{
			name:   "no update",
			runner: &RunnerMeta{Status: RunnerIdle, LastPingAt: now},
			want:   "",
		},
		{
			name:   "update from idle to unknown",
			runner: &RunnerMeta{Status: RunnerIdle, LastPingAt: now.Add(-pingTimeout).Add(-time.Second)},
			want:   RunnerUnknown,
		},
		{
			name:   "update from unknown to errored",
			runner: &RunnerMeta{Status: RunnerUnknown, LastStatusAt: now.Add(-6 * time.Minute)},
			want:   RunnerErrored,
		},
		{
			name:        "delete",
			runner:      &RunnerMeta{Status: RunnerErrored, LastStatusAt: now.Add(-2 * time.Hour)},
			want:        "",
			wantDeleted: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{}
			m := &manager{client: svc}
			err := m.update(context.Background(), tt.runner)
			require.NoError(t, err)
			assert.Equal(t, tt.want, svc.status)
		})
	}
}
