// Package agent contains code related to agents
package agent

import (
	"net"
	"time"
)

type Status string

const (
	AgentIdle    Status = "idle"
	AgentBusy    Status = "busy"
	AgentExited  Status = "exited"
	AgentErrored Status = "errored"
)

type Agent struct {
	// Unique system-wide ID
	ID string
	// Optional name
	Name *string
	// Current status of agent
	Status Status
	// Whether it is built into otfd (true) or is a separate otf-agent process
	// (false)
	Server bool
	// Last time a ping was received from the agent
	LastPingAt time.Time
	// IP address of agent
	IPAddress net.IPAddr
}
