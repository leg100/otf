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
	AgentID string
}

// PhaseFinishOptions report the status of a phase upon finishing.
type PhaseFinishOptions struct {
	// Errored is true if the phase finished unsuccessfully.
	Errored bool
}

type PhaseStatusTimestamp struct {
	Status    PhaseStatus
	Timestamp time.Time
}

// phaseStatus is a mixin providing status functionality for a phase
type phaseStatus struct {
	status           PhaseStatus
	statusTimestamps []PhaseStatusTimestamp
}

func (p *phaseStatus) Status() PhaseStatus                      { return p.status }
func (p *phaseStatus) StatusTimestamps() []PhaseStatusTimestamp { return p.statusTimestamps }

func (p *phaseStatus) StatusTimestamp(status PhaseStatus) (time.Time, error) {
	for _, rst := range p.statusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, ErrStatusTimestampNotFound
}

func (p *phaseStatus) updateStatus(status PhaseStatus) {
	p.status = status
	p.statusTimestamps = append(p.statusTimestamps, PhaseStatusTimestamp{
		Status:    status,
		Timestamp: CurrentTimestamp(),
	})
}

func newPhaseStatus() *phaseStatus {
	p := &phaseStatus{}
	p.updateStatus(PhasePending)
	return p
}
