package sqlite

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_Create(t *testing.T) {
	db := newTestDB(t)
	ws := createTestWorkspace(t, db, "ws-123", "default")
	cv := createTestConfigurationVersion(t, db, "cv-123", ws)

	rdb := NewRunDB(db)

	run, err := rdb.Create(newTestRun("run-123", ws, cv))
	require.NoError(t, err)

	// Ensure primary keys populated
	assert.Equal(t, int64(1), run.Model.ID)
	assert.Equal(t, int64(1), run.Plan.Model.ID)
	assert.Equal(t, int64(1), run.Apply.Model.ID)

	// Ensure foreign keys populated
	assert.Equal(t, int64(1), run.Plan.RunID)
	assert.Equal(t, int64(1), run.Apply.RunID)
	assert.Equal(t, int64(1), run.Workspace.Model.ID)
	assert.Equal(t, int64(1), run.ConfigurationVersion.Model.ID)
}

func TestRun_Update(t *testing.T) {
	db := newTestDB(t)
	ws := createTestWorkspace(t, db, "ws-123", "default")
	cv := createTestConfigurationVersion(t, db, "cv-123", ws)
	run := createTestRun(t, db, "run-123", ws, cv)

	rdb := NewRunDB(db)

	updated, err := rdb.Update(run.ID, func(run *otf.Run) error {
		run.Status = otf.RunPlanQueued
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, otf.RunPlanQueued, updated.Status)
}

func TestRun_Get(t *testing.T) {
	db := newTestDB(t)
	ws := createTestWorkspace(t, db, "ws-123", "default")
	cv := createTestConfigurationVersion(t, db, "cv-123", ws)
	run := createTestRun(t, db, "run-123", ws, cv)

	rdb := NewRunDB(db)

	_, err := rdb.Get(otf.RunGetOptions{ID: otf.String(run.ID)})
	require.NoError(t, err)
}

func TestRun_List(t *testing.T) {
	tests := []struct {
		name string
		opts otf.RunListOptions
		want int
	}{
		{
			name: "default",
			opts: otf.RunListOptions{},
			want: 2,
		},
		{
			name: "filter by workspace",
			opts: otf.RunListOptions{WorkspaceID: otf.String("ws-123")},
			want: 1,
		},
		{
			name: "filter by status",
			opts: otf.RunListOptions{Statuses: []otf.RunStatus{otf.RunPending}},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			ws1 := createTestWorkspace(t, db, "ws-123", "dev")
			ws2 := createTestWorkspace(t, db, "ws-345", "prod")
			cv1 := createTestConfigurationVersion(t, db, "cv-123", ws1)
			cv2 := createTestConfigurationVersion(t, db, "cv-345", ws2)
			_ = createTestRun(t, db, "run-123", ws1, cv1)
			_ = createTestRun(t, db, "run-345", ws2, cv2)

			rdb := NewRunDB(db)

			results, err := rdb.List(tt.opts)
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(results.Items))
		})
	}
}
