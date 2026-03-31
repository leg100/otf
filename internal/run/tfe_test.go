package run

import (
	"testing"

	"github.com/leg100/otf/internal/runstatus"
	"github.com/stretchr/testify/require"
)

func TestTFEStatus(t *testing.T) {
	api := &tfe{}

	require.Equal(t, string(runstatus.CostEstimated), api.tfeStatus(runstatus.PolicyChecked))
	require.Equal(t, string(runstatus.CostEstimated), api.tfeStatus(runstatus.PolicySoftFailed))
	require.Equal(t, string(runstatus.PolicyFailed), api.tfeStatus(runstatus.PolicyFailed))
}
