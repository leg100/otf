package runstatus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlannedCompatible(t *testing.T) {
	assert.True(t, PlannedCompatible(Planned))
	assert.True(t, PlannedCompatible(CostEstimated))
	assert.True(t, PlannedCompatible(PolicyChecked))
	assert.True(t, PlannedCompatible(PolicySoftFailed))
	assert.False(t, PlannedCompatible(Applied))
}

func TestExpandPlannedCompatible(t *testing.T) {
	assert.Equal(t,
		[]Status{Planned, CostEstimated, PolicyChecked, PolicySoftFailed},
		ExpandPlannedCompatible([]Status{Planned}),
	)

	assert.Equal(t,
		[]Status{Planned, CostEstimated, PolicyChecked, PolicySoftFailed, Applied},
		ExpandPlannedCompatible([]Status{Planned, Applied, PolicyChecked}),
	)

	assert.Nil(t, ExpandPlannedCompatible(nil))
}
