// Package mode handles the various execution modes for workspace runs.
package mode

import (
	"errors"
	"fmt"

	"github.com/leg100/otf/internal/resource"
)

type Mode string

const (
	// Remote mode executes the run via the otf server, otfd.
	Remote Mode = "remote"
	// Local mode executes the run on the user's local machine
	Local Mode = "local"
	// Agent mode executes the run via the agent, otf-agent.
	Agent Mode = "agent"
)

var (
	ErrAgentExecutionModeWithoutPool = errors.New("agent execution mode requires agent pool ID")
	ErrNonAgentExecutionModeWithPool = errors.New("agent pool ID can only be specified with agent execution mode")
)

// Validate the execution mode and optionally the agent pool ID. The two options
// are intimately related, hence the validation of the parameters in tandem.
func Validate(mode Mode, agentPoolID *resource.TfeID) error {
	switch mode {
	case Agent:
		if agentPoolID == nil {
			return ErrAgentExecutionModeWithoutPool
		}
	case Remote, Local:
		// No pool ID should be provided
		if agentPoolID != nil {
			return ErrNonAgentExecutionModeWithPool
		}
	default:
		return fmt.Errorf("invalid execution mode: %s", mode)
	}
	return nil
}
