package agent

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
