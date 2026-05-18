// Package mode handles the various execution modes for workspace runs.
package mode

type Mode string

const (
	// Remote mode executes the run via the otf server, otfd.
	Remote Mode = "remote"
	// Local mode executes the run on the user's local machine
	Local Mode = "local"
	// Agent mode executes the run via the agent, otf-agent.
	Agent Mode = "agent"
)
