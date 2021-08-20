package agent

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRunLister struct {
	runs []*ots.Run
}

func (l *mockRunLister) List(opts ots.RunListOptions) (*ots.RunList, error) {
	return &ots.RunList{Items: l.runs}, nil
}

// TestNewSpooler tests the spooler constructor
func TestNewSpooler(t *testing.T) {
	want := &ots.Run{ID: "run-123", Status: tfe.RunPlanQueued}

	spooler, err := NewSpooler(
		&mockRunLister{runs: []*ots.Run{want}},
		&mock.EventService{},
		logr.Discard(),
	)
	require.NoError(t, err)

	assert.Equal(t, want, <-spooler.queue)
}
