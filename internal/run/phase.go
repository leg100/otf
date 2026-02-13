package run

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

const (
	PhasePending     PhaseStatus = "pending"
	PhaseQueued      PhaseStatus = "queued"
	PhaseRunning     PhaseStatus = "running"
	PhaseFinished    PhaseStatus = "finished"
	PhaseCanceled    PhaseStatus = "canceled"
	PhaseErrored     PhaseStatus = "errored"
	PhaseUnreachable PhaseStatus = "unreachable"

	PendingPhase PhaseType = "pending"
	PlanPhase    PhaseType = "plan"
	ApplyPhase   PhaseType = "apply"
	FinalPhase   PhaseType = "final"
	UnknownPhase PhaseType = "unknown"
)

type (
	// Phase is a section of work performed by a run.
	Phase struct {
		RunID  resource.TfeID `json:"run_id"`
		Status PhaseStatus    `json:"status"`

		// Timestamps of when a state transition occured. Ordered earliest
		// first.
		StatusTimestamps []PhaseStatusTimestamp `json:"status_timestamps"`

		PhaseType `json:"phase"`

		// report of planned or applied resource changes
		ResourceReport *Report `json:"resource_report"`
		// report of planned or applied output changes
		OutputReport *Report `json:"output_report"`
	}

	PhaseStatus string

	PhaseType string

	PhaseStartOptions struct {
		Type    string         `jsonapi:"primary,phase"`
		AgentID resource.TfeID `jsonapi:"attribute" json:"agent-id"`
	}

	// PhaseFinishOptions report the status of a phase upon finishing.
	PhaseFinishOptions struct {
		Errored bool `json:"errored,omitempty"`
	}

	PhaseStatusTimestamp struct {
		Phase     PhaseType
		Status    PhaseStatus `json:"status"`
		Timestamp time.Time   `json:"timestamp"`
	}
)

// newPhase constructs a new phase. A new phase always starts in pending status.
func newPhase(runID resource.TfeID, t PhaseType) Phase {
	p := Phase{RunID: runID, PhaseType: t}
	p.UpdateStatus(PhasePending)
	return p
}

func (p *Phase) HasChanges() bool {
	var (
		hasResourceChanges = p.ResourceReport != nil && p.ResourceReport.HasChanges()
		hasOutputChanges   = p.OutputReport != nil && p.OutputReport.HasChanges()
	)
	return hasResourceChanges || hasOutputChanges
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
		Phase:     p.PhaseType,
		Status:    status,
		Timestamp: internal.CurrentTimestamp(nil),
	})
}

func (p *Phase) HasStarted() bool {
	_, err := p.StatusTimestamp(PhaseRunning)
	return err == nil
}

// StartedAt returns the time the phase started running, returning zero time
// if it is yet to start running.
func (p *Phase) StartedAt() time.Time {
	start, err := p.StatusTimestamp(PhaseRunning)
	if err != nil {
		// yet to enter running state
		return time.Time{}
	}
	return start
}

// ElapsedTime returns the time taken for the phase to complete its running
// state. If the run is yet to enter a running state then it returns 0. If the
// running state is still in progress then it returns the time since entering
// the running state.
func (p *Phase) ElapsedTime(now time.Time) time.Duration {
	start, err := p.StatusTimestamp(PhaseRunning)
	if err != nil {
		// yet to enter running state
		return 0
	}
	// lazily look for another timestamp later than that given for the running
	// state; if such a timestamp is found then return the difference between
	// that and the time the running state was entered; otherwise assume still
	// running
	for _, st := range p.StatusTimestamps {
		if st.Timestamp.After(start) {
			return st.Timestamp.Sub(start)
		}
	}
	// still running
	return now.Sub(start)
}

func (p *Phase) Done() bool {
	switch p.Status {
	case PhaseFinished, PhaseCanceled, PhaseErrored, PhaseUnreachable:
		return true
	default:
		return false
	}
}

func (p *Phase) String() string { return string(p.PhaseType) }

func (s PhaseStatus) String() string { return string(s) }
