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

type (
	// Phase is a section of work performed by a run.
	Phase struct {
		RunID string

		PhaseType // plan or apply

		*ResourceReport // report of planned or applied resource changes

		Status           PhaseStatus // current phase status
		StatusTimestamps []PhaseStatusTimestamp
	}

	PhaseType   string
	PhaseStatus string

	PhaseStartOptions struct {
		AgentID string `jsonapi:"attr,agent-id,omitempty"`
	}

	// PhaseFinishOptions report the status of a phase upon finishing.
	PhaseFinishOptions struct {
		// Errored is true if the phase finished unsuccessfully.
		Errored bool `jsonapi:"attr,errored,omitempty"`
	}

	PhaseStatusTimestamp struct {
		Status    PhaseStatus
		Timestamp time.Time
	}
)

// NewPhase constructs a new phase. A new phase always starts in pending status.
func NewPhase(runID string, t PhaseType) Phase {
	p := Phase{RunID: runID, PhaseType: t}
	p.UpdateStatus(PhasePending)
	return p
}

// Lookup timestamp for status
func (p *Phase) StatusTimestamp(status PhaseStatus) (time.Time, error) {
	for _, rst := range p.StatusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, ErrStatusTimestampNotFound
}

func (p *Phase) UpdateStatus(status PhaseStatus) {
	p.Status = status
	p.StatusTimestamps = append(p.StatusTimestamps, PhaseStatusTimestamp{
		Status:    status,
		Timestamp: CurrentTimestamp(),
	})
}

func (s PhaseStatus) String() string { return string(s) }
