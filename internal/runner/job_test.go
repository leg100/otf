package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_jobSpecFromString(t *testing.T) {
	tests := []struct {
		spec    string
		want    JobSpec
		wantErr error
	}{
		{
			spec: "run-grDQCjrQne1EUIGW/plan",
			want: JobSpec{RunID: "run-grDQCjrQne1EUIGW", Phase: "plan"},
		},
		{
			spec:    "grDQCjrQne1EUIGW/plan",
			wantErr: ErrMalformedJobSpecString,
		},
		{
			spec:    "run-grDQCjrQne1EUIGW",
			wantErr: ErrMalformedJobSpecString,
		},
	}
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			got, err := jobSpecFromString(tt.spec)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

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
