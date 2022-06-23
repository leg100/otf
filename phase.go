package otf

import (
	"context"
	"errors"
	"time"
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

var (
	ErrPhaseAlreadyStarted = errors.New("phase already started")
)

type PhaseStatus string

// Phase is a section of work performed by a run.
type Phase interface {
	// Do some work in an execution environment
	Do(Environment) error
	// GetID gets the ID of the Job
	PhaseID() string
	PhaseStatus() PhaseStatus
	// Get job status timestamps
	PhaseStatusTimestamps() []PhaseStatusTimestamp
	PhaseStatusTimestamp(PhaseStatus) (time.Time, error)
	// Service provides an appropriate application service to interact with the
	// phase
	Service(Application) (PhaseService, error)
}

type PhaseService interface {
	// Start a phase. ErrJobAlreadyStarted is returned if phase has already been
	// started.
	Start(ctx context.Context, id string, opts PhaseStartOptions) (*Run, error)
	// Finish is called by an agent when it finishes a job.
	Finish(ctx context.Context, id string, opts PhaseFinishOptions) (*Run, error)
	// Retrieve and upload chunks of logs for jobs
	ChunkService
}

type PhaseStartOptions struct {
	AgentID string
}

type PhaseFinishOptions struct {
	Errored bool
}

type PhaseStatusTimestamp struct {
	Status    PhaseStatus
	Timestamp time.Time
}

type phaseMixin struct {
	status           PhaseStatus
	statusTimestamps []PhaseStatusTimestamp
}

func (p *phaseMixin) Status() PhaseStatus                           { return p.status }
func (p *phaseMixin) StatusTimestamps() []PhaseStatusTimestamp      { return p.statusTimestamps }
func (p *phaseMixin) PhaseStatus() PhaseStatus                      { return p.status }
func (p *phaseMixin) PhaseStatusTimestamps() []PhaseStatusTimestamp { return p.statusTimestamps }

func (p *phaseMixin) PhaseStatusTimestamp(status PhaseStatus) (time.Time, error) {
	for _, rst := range p.statusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, ErrStatusTimestampNotFound
}

func (p *phaseMixin) updateStatus(status PhaseStatus) {
	p.status = status
	p.statusTimestamps = append(p.statusTimestamps, PhaseStatusTimestamp{
		Status:    status,
		Timestamp: CurrentTimestamp(),
	})
}

func newPhase() *phaseMixin {
	p := &phaseMixin{
		status: PhasePending,
	}
	return p
}

// pendingPhase is the initial phase of a run.
type pendingPhase struct {
	Phase
}
