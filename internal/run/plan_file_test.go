package run

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanFile(t *testing.T) {
	data, err := os.ReadFile("testdata/plan.json")
	require.NoError(t, err)

	file := PlanFile{}
	require.NoError(t, json.Unmarshal(data, &file))

	want := PlanFile{
		ResourceChanges: []ResourceChange{
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
		OutputChanges: map[string]Change{
			"random_string": {
				Actions: []ChangeAction{
					CreateAction,
				},
			},
		},
	}
	assert.Equal(t, want, file)
}

func TestPlanFile_Changes(t *testing.T) {
	data, err := os.ReadFile("testdata/plan.json")
	require.NoError(t, err)

	file := PlanFile{}
	require.NoError(t, json.Unmarshal(data, &file))

	resourceReport, outputReport := file.Summarize()

	assert.Equal(t, 2, resourceReport.Additions)
	assert.Equal(t, 0, resourceReport.Changes)
	assert.Equal(t, 0, resourceReport.Destructions)

	assert.Equal(t, 1, outputReport.Additions)
	assert.Equal(t, 0, outputReport.Changes)
	assert.Equal(t, 0, outputReport.Destructions)
}
