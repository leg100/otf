// Package execution handles the various kinds of execution for workspace runs.
package execution

import (
	"errors"
	"fmt"

	"github.com/leg100/otf/internal/resource"
)

// Kind of execution for workspace run.
type Kind string

// Mode of execution for workspace run, combining execution kind and any
// additional data relating to kind.
type Mode struct {
	kind        Kind
	agentPoolID *resource.TfeID
}

const (
	// RemoteKind executes the run via the otf server, otfd.
	RemoteKind Kind = "remote"
	// LocalKind executes the run on the user's local machine
	LocalKind Kind = "local"
	// AgentKind executes the run via the agent, otf-agent.
	AgentKind Kind = "agent"
)

var (
	ErrAgentExecutionModeWithoutPool = errors.New("agent execution mode requires agent pool ID")
	ErrNonAgentExecutionModeWithPool = errors.New("agent pool ID can only be specified with agent execution mode")
)

func NewMode(kind Kind, agentPoolID *resource.TfeID) (Mode, error) {
	if err := isValid(kind, agentPoolID); err != nil {
		return Mode{}, err
	}

	return Mode{kind: kind, agentPoolID: agentPoolID}, nil
}

func NewModeWithDefaults(overrideKind *Kind, defaultKind Kind, overrideAgentPoolID, defaultAgentPoolID *resource.TfeID) (Mode, error) {
	// Use default kind if no override kind provided.
	if overrideKind == nil {
		overrideKind = &defaultKind
	}
	// If agent kind and no agent pool ID override provided then use default
	// agent pool ID.
	if *overrideKind == AgentKind && overrideAgentPoolID == nil {
		overrideAgentPoolID = defaultAgentPoolID
	}

	return NewMode(*overrideKind, overrideAgentPoolID)
}

func RemoteMode() Mode                     { return Mode{kind: RemoteKind} }
func LocalMode() Mode                      { return Mode{kind: LocalKind} }
func AgentMode(poolID resource.TfeID) Mode { return Mode{kind: AgentKind, agentPoolID: &poolID} }

// Update the mode and/or agent pool ID. An error is returned if an invalid
// combination is provided. True is returned if something was set.
func (m *Mode) Update(kind *Kind, agentPoolID *resource.TfeID) error {
	if kind == nil && agentPoolID == nil {
		// Nothing to set.
		return nil
	}

	if kind != nil {
		m.kind = *kind
	}
	m.agentPoolID = agentPoolID

	if err := isValid(m.kind, agentPoolID); err != nil {
		return err
	}

	return nil
}

func (m Mode) Kind() Kind                   { return m.kind }
func (m Mode) AgentPoolID() *resource.TfeID { return m.agentPoolID }

func isValid(kind Kind, agentPoolID *resource.TfeID) error {
	switch kind {
	case LocalKind, RemoteKind:
		if agentPoolID != nil {
			// Non-agent modes are mutually exclusive with agent pool ID
			return ErrNonAgentExecutionModeWithPool
		}
	case AgentKind:
		if agentPoolID == nil {
			// Agent mode requires agent pool ID
			return ErrAgentExecutionModeWithoutPool
		}
	default:
		return fmt.Errorf("invalid kind: %s", kind)
	}
	return nil
}
