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

	assert.Equal(t, int64(1), run.Model.ID)
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
