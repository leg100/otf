package inmem

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceQueue(t *testing.T) {
	t.Run("skip speculative run", func(t *testing.T) {
		q := NewWorkspaceQueue()
		speculative := otf.NewTestRun(t, otf.TestRunCreateOptions{Speculative: true})
		q.Update(speculative)
		assert.Equal(t, 0, len(q.Get()))
	})

	t.Run("add run", func(t *testing.T) {
		q := NewWorkspaceQueue()
		run := otf.NewTestRun(t, otf.TestRunCreateOptions{})
		q.Update(run)
		assert.Equal(t, 1, len(q.Get()))
	})

	t.Run("update run", func(t *testing.T) {
		q := NewWorkspaceQueue()
		run := otf.NewTestRun(t, otf.TestRunCreateOptions{})
		q.Update(run)
		assert.Equal(t, run, q.Get()[0])

		assert.NoError(t, run.EnqueuePlan(context.Background(), &otf.FakeLatestRunSetter{}))
		q.Update(run)
		assert.Equal(t, run, q.Get()[0])
	})

	t.Run("remove run", func(t *testing.T) {
		q := NewWorkspaceQueue()
		run := otf.NewTestRun(t, otf.TestRunCreateOptions{})
		q.Update(run)
		assert.Equal(t, 1, len(q.Get()))
		run.Discard()
		q.Update(run)
		assert.Equal(t, 0, len(q.Get()))
	})
}
