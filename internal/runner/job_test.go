package runner

import (
	"testing"
	"time"

	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/stretchr/testify/assert"
)

func TestJob_updateStatus(t *testing.T) {
	tests := []struct {
		name string
		from JobStatus
		to   JobStatus
		want error
	}{
		{"allocate job", JobUnallocated, JobAllocated, nil},
		{"start job", JobAllocated, JobRunning, nil},
		{"finish job", JobRunning, JobFinished, nil},
		{"finish with error", JobRunning, JobErrored, nil},
		{"cancel unstarted job", JobAllocated, JobCanceled, nil},
		{"cancel running job", JobRunning, JobCanceled, nil},
		{"cannot allocate canceled job", JobCanceled, JobAllocated, ErrInvalidJobStateTransition},
		{"cannot allocate finished job", JobCanceled, JobFinished, ErrInvalidJobStateTransition},
		{"cannot allocate errored job", JobCanceled, JobErrored, ErrInvalidJobStateTransition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &Job{Status: tt.from}
			assert.Equal(t, tt.want, j.updateStatus(tt.to))
		})
	}
}

func TestJob_cancel(t *testing.T) {
	tests := []struct {
		name       string
		status     JobStatus
		run        *run.Run
		wantStatus JobStatus
		wantSignal bool
		wantForce  bool
		wantError  error
	}{
		{
			name:       "cancel active job",
			status:     JobRunning,
			run:        &run.Run{Status: runstatus.Planning, CancelSignaledAt: new(time.Now())},
			wantStatus: JobRunning,
			wantSignal: true,
			wantForce:  false,
			wantError:  nil,
		},
		{
			name:       "force cancel active job",
			status:     JobRunning,
			run:        &run.Run{Status: runstatus.ForceCanceled},
			wantStatus: JobCanceled,
			wantSignal: true,
			wantForce:  true,
			wantError:  nil,
		},
		{
			name:       "cancel job but dont signal it",
			status:     JobRunning,
			run:        &run.Run{Status: runstatus.Canceled},
			wantStatus: JobCanceled,
			wantSignal: false,
			wantForce:  false,
			wantError:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{Status: tt.status}
			gotSignal, gotForce, gotError := job.cancel(tt.run)
			assert.Equal(t, tt.wantStatus, job.Status)
			assert.Equal(t, tt.wantSignal, gotSignal)
			assert.Equal(t, tt.wantForce, gotForce)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}
