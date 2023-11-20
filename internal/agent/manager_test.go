package agent

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
		agent       *Agent
		want        AgentStatus
		wantDeleted bool
	}{
		{
			name:  "no update",
			agent: &Agent{Status: AgentIdle, LastPingAt: now},
			want:  "",
		},
		{
			name:  "update from idle to unknown",
			agent: &Agent{Status: AgentIdle, LastPingAt: now.Add(-pingTimeout).Add(-time.Second)},
			want:  AgentUnknown,
		},
		{
			name:  "update from unknown to errored",
			agent: &Agent{Status: AgentUnknown, LastStatusAt: now.Add(-6 * time.Minute)},
			want:  AgentErrored,
		},
		{
			name:        "delete",
			agent:       &Agent{Status: AgentErrored, LastStatusAt: now.Add(-2 * time.Hour)},
			want:        "",
			wantDeleted: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{}
			m := &manager{Service: svc}
			err := m.update(context.Background(), tt.agent)
			require.NoError(t, err)
			assert.Equal(t, tt.want, svc.status)
		})
	}
}
