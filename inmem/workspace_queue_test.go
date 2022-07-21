package inmem

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceQueue(t *testing.T) {
	ctx := context.Background()

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

		assert.NoError(t, run.EnqueuePlan(ctx, &otf.FakeLatestRunSetter{}))
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

	t.Run("multiple operations", func(t *testing.T) {
		q := NewWorkspaceQueue()
		run1 := otf.NewTestRun(t, otf.TestRunCreateOptions{})
		run2 := otf.NewTestRun(t, otf.TestRunCreateOptions{})
		run3 := otf.NewTestRun(t, otf.TestRunCreateOptions{})
		run4 := otf.NewTestRun(t, otf.TestRunCreateOptions{})

		q.Update(run1)
		q.Update(run2)
		q.Update(run3)

		assert.Equal(t, 3, len(q.Get()))

		run2.Discard()
		q.Update(run2)

		assert.Equal(t, 2, len(q.Get()))

		err := run1.EnqueuePlan(ctx, &otf.FakeLatestRunSetter{})
		assert.NoError(t, err)
		q.Update(run1)

		_, err = run3.Cancel()
		assert.NoError(t, err)
		q.Update(run3)

		assert.Equal(t, 1, len(q.Get()))

		q.Update(run4)

		assert.Equal(t, 2, len(q.Get()))
	})
}
