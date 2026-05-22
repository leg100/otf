package execution

import (
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPoolID() resource.TfeID {
	return resource.NewTfeID(resource.AgentPoolKind)
}

func TestNew(t *testing.T) {
	poolID := newPoolID()
	tests := []struct {
		name        string
		kind        Kind
		agentPoolID *resource.TfeID
		want        Mode
		wantErr     error
	}{
		{
			name: "remote without pool",
			kind: RemoteKind,
			want: RemoteMode(),
		},
		{
			name: "local without pool",
			kind: LocalKind,
			want: LocalMode(),
		},
		{
			name:        "agent with pool",
			kind:        AgentKind,
			agentPoolID: &poolID,
			want:        AgentMode(poolID),
		},
		{
			name:    "agent without pool",
			kind:    AgentKind,
			wantErr: ErrAgentExecutionModeWithoutPool,
		},
		{
			name:        "remote with pool",
			kind:        RemoteKind,
			agentPoolID: &poolID,
			wantErr:     ErrNonAgentExecutionModeWithPool,
		},
		{
			name:        "local with pool",
			kind:        LocalKind,
			agentPoolID: &poolID,
			wantErr:     ErrNonAgentExecutionModeWithPool,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewMode(tc.kind, tc.agentPoolID)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewWithDefaults(t *testing.T) {
	poolID := newPoolID()
	tests := []struct {
		name               string
		kind               *Kind
		defaultKind        Kind
		agentPoolID        *resource.TfeID
		defaultAgentPoolID *resource.TfeID
		want               Mode
		wantErr            error
	}{
		{
			name:        "uses defaults when both nil",
			defaultKind: RemoteKind,
			want:        RemoteMode(),
		},
		{
			name:        "override kind",
			kind:        new(LocalKind),
			defaultKind: RemoteKind,
			want:        LocalMode(),
		},
		{
			name:               "override pool",
			defaultKind:        AgentKind,
			defaultAgentPoolID: &poolID,
			agentPoolID:        &poolID,
			want:               AgentMode(poolID),
		},
		{
			name:        "override to agent with pool",
			kind:        new(AgentKind),
			defaultKind: RemoteKind,
			agentPoolID: &poolID,
			want:        AgentMode(poolID),
		},
		{
			name:        "agent default without pool",
			defaultKind: AgentKind,
			wantErr:     ErrAgentExecutionModeWithoutPool,
		},
		{
			name:        "override to agent without pool",
			kind:        new(AgentKind),
			defaultKind: RemoteKind,
			wantErr:     ErrAgentExecutionModeWithoutPool,
		},
		{
			name:               "override to remote with default pool still set",
			kind:               new(RemoteKind),
			defaultKind:        AgentKind,
			defaultAgentPoolID: &poolID,
			want:               RemoteMode(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewModeWithDefaults(tc.kind, tc.defaultKind, tc.agentPoolID, tc.defaultAgentPoolID)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUpdate(t *testing.T) {
	poolID := newPoolID()
	tests := []struct {
		name        string
		initial     Mode
		kind        *Kind
		agentPoolID *resource.TfeID
		wantKind    Kind
		wantPool    *resource.TfeID
		wantErr     error
	}{
		{
			name:     "no-op when both nil",
			initial:  RemoteMode(),
			wantKind: RemoteKind,
		},
		{
			name:     "set kind only",
			initial:  RemoteMode(),
			kind:     new(LocalKind),
			wantKind: LocalKind,
		},
		{
			name:        "set pool only",
			initial:     AgentMode(poolID),
			agentPoolID: &poolID,
			wantKind:    AgentKind,
			wantPool:    &poolID,
		},
		{
			name:        "set both kind and pool",
			initial:     RemoteMode(),
			kind:        new(AgentKind),
			agentPoolID: &poolID,
			wantKind:    AgentKind,
			wantPool:    &poolID,
		},
		{
			name:        "set to agent kind without pool",
			initial:     RemoteMode(),
			kind:        new(AgentKind),
			agentPoolID: nil,
			wantKind:    AgentKind,
			wantPool:    &poolID,
			wantErr:     ErrAgentExecutionModeWithoutPool,
		},
		{
			name:        "set to remote kind with pool",
			initial:     LocalMode(),
			kind:        new(RemoteKind),
			agentPoolID: &poolID,
			wantKind:    AgentKind,
			wantPool:    &poolID,
			wantErr:     ErrNonAgentExecutionModeWithPool,
		},
		{
			name:     "overwrite kind",
			initial:  LocalMode(),
			kind:     new(RemoteKind),
			wantKind: RemoteKind,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.initial
			err := m.Update(tc.kind, tc.agentPoolID)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantKind, m.Kind())
			assert.Equal(t, tc.wantPool, m.AgentPoolID())
		})
	}
}
