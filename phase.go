package otf

import (
	"errors"
	"time"
)

const (
	PendingPhase PhaseType = "pending"
	PlanPhase    PhaseType = "plan"
	ApplyPhase   PhaseType = "apply"
	FinalPhase   PhaseType = "final"
	UnknownPhase PhaseType = "unknown"

	PhasePending     PhaseStatus = "pending"
	PhaseQueued      PhaseStatus = "queued"
	PhaseRunning     PhaseStatus = "running"
	PhaseFinished    PhaseStatus = "finished"
	PhaseCanceled    PhaseStatus = "canceled"
	PhaseErrored     PhaseStatus = "errored"
	PhaseUnreachable PhaseStatus = "unreachable"
)

var ErrPhaseAlreadyStarted = errors.New("phase already started")

// PhaseSpec specifies a phase of a run
type PhaseSpec struct {
	RunID string
	Phase PhaseType
}

type PhaseStatus string

func (r PhaseStatus) String() string { return string(r) }

// Phase is a section of work performed by a run.
type Phase interface {
	// Run ID
	ID() string
	// phase type
	Phase() PhaseType
	// current phase status
	Status() PhaseStatus
	// Get job status timestamps
	StatusTimestamps() []PhaseStatusTimestamp
	// Lookup timestamp for status
	StatusTimestamp(PhaseStatus) (time.Time, error)
}

type PhaseType string

type PhaseStartOptions struct {
	AgentID string `jsonapi:"attr,agent-id,omitempty"`
}

// PhaseFinishOptions report the status of a phase upon finishing.
type PhaseFinishOptions struct {
	// Errored is true if the phase finished unsuccessfully.
	Errored bool `jsonapi:"attr,errored,omitempty"`
}

type PhaseStatusTimestamp struct {
	Status    PhaseStatus
	Timestamp time.Time
}
