package sqlite

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_Create(t *testing.T) {
	db := NewRunDB(newTestDB(t))

	run, err := db.Create(newTestRun())
	require.NoError(t, err)

	assert.Equal(t, int64(1), run.Model.ID)
}

func TestRun_Get(t *testing.T) {
	db := newTestDB(t)

	odb := NewOrganizationDB(db)
	rdb := NewRunDB(db)

	_, err := odb.Create(newTestOrganization("org-123"))
	require.NoError(t, err)

	testRun := newTestRun()

	_, err = rdb.Create(testRun)
	require.NoError(t, err)

	_, err = rdb.Get(otf.RunGetOptions{ID: otf.String(testRun.ID)})
	require.NoError(t, err)
}
