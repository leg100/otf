package run

import (
	"errors"
	"time"

	"github.com/leg100/otf/internal"
)

const (
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

		internal.PhaseType
		*ResourceReport // report of planned or applied resource changes

		Status           PhaseStatus // current phase status
		StatusTimestamps []PhaseStatusTimestamp
	}

	PhaseStatus string

	PhaseStartOptions struct {
		Type    string `jsonapi:"primary,phase"`
		AgentID string `jsonapi:"attribute" json:"agent-id,omitempty"`
	}

	// PhaseFinishOptions report the status of a phase upon finishing.
	PhaseFinishOptions struct {
		Type string `jsonapi:"primary,phase"`
		// Errored is true if the phase finished unsuccessfully.
		Errored bool `jsonapi:"attribute" json:"errored,omitempty"`
	}

	PhaseStatusTimestamp struct {
		Status    PhaseStatus
		Timestamp time.Time
	}
)

// NewPhase constructs a new phase. A new phase always starts in pending status.
func NewPhase(runID string, t internal.PhaseType) Phase {
	p := Phase{RunID: runID, PhaseType: t}
	p.UpdateStatus(PhasePending)
	return p
}

func (p *Phase) HasChanges() bool {
	if p.ResourceReport != nil {
		return p.ResourceReport.HasChanges()
	}
	// no report has been published yet, which means there are no proposed
	// changes yet.
	return false
}

// StatusTimestamp looks up the timestamp for a status
func (p *Phase) StatusTimestamp(status PhaseStatus) (time.Time, error) {
	for _, rst := range p.StatusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, internal.ErrStatusTimestampNotFound
}

func (p *Phase) UpdateStatus(status PhaseStatus) {
	p.Status = status
	p.StatusTimestamps = append(p.StatusTimestamps, PhaseStatusTimestamp{
		Status:    status,
		Timestamp: internal.CurrentTimestamp(),
	})
}

func (s PhaseStatus) String() string { return string(s) }
