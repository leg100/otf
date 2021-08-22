package ots

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanFile(t *testing.T) {
	data, err := os.ReadFile("agent/testdata/plan.json")
	require.NoError(t, err)

	file := PlanFile{}
	require.NoError(t, json.Unmarshal(data, &file))

	want := PlanFile{
		ResourcesChanges: []ResourceChange{
			{
				Change: Change{
					Actions: []ChangeAction{
						CreateAction,
					},
				},
			},
			{
				Change: Change{
					Actions: []ChangeAction{
						CreateAction,
					},
				},
			},
		},
	}
	assert.Equal(t, want, file)
}

func TestPlanFile_Changes(t *testing.T) {
	data, err := os.ReadFile("agent/testdata/plan.json")
	require.NoError(t, err)

	file := PlanFile{}
	require.NoError(t, json.Unmarshal(data, &file))

	adds, updates, deletes := file.Changes()

	assert.Equal(t, 2, adds)
	assert.Equal(t, 0, updates)
	assert.Equal(t, 0, deletes)
}
