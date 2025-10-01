package run

import (
	"strings"
	"testing"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMetricsCollector_bootstrap(t *testing.T) {
	runStatusMetric.Reset()

	const metadata = `
		# HELP otf_runs_statuses Total runs by status
		# TYPE otf_runs_statuses gauge
`

	mc := &MetricsCollector{}
	mc.bootstrap(
		status{ID: testutils.ParseID(t, "run-1"), Status: runstatus.Pending},
		status{ID: testutils.ParseID(t, "run-2"), Status: runstatus.Pending},
		status{ID: testutils.ParseID(t, "run-3"), Status: runstatus.Pending},
		status{ID: testutils.ParseID(t, "run-4"), Status: runstatus.Pending},
		status{ID: testutils.ParseID(t, "run-5"), Status: runstatus.Planning},
		status{ID: testutils.ParseID(t, "run-6"), Status: runstatus.Planning},
		status{ID: testutils.ParseID(t, "run-7"), Status: runstatus.Planning},
		status{ID: testutils.ParseID(t, "run-8"), Status: runstatus.Planning},
		status{ID: testutils.ParseID(t, "run-9"), Status: runstatus.Applied},
		status{ID: testutils.ParseID(t, "run-10"), Status: runstatus.Applied},
		status{ID: testutils.ParseID(t, "run-11"), Status: runstatus.Applied},
		status{ID: testutils.ParseID(t, "run-12"), Status: runstatus.Applied},
	)
	assert.Len(t, mc.currentStatuses, 12)
	want := `
		otf_runs_statuses{status="applied"} 4
		otf_runs_statuses{status="pending"} 4
		otf_runs_statuses{status="planning"} 4
	`
	assert.NoError(t, testutil.CollectAndCompare(runStatusMetric, strings.NewReader(metadata+want), "otf_runs_statuses"))
}

func TestMetricsCollector_update(t *testing.T) {
	runStatusMetric.Reset()

	const metadata = `
		# HELP otf_runs_statuses Total runs by status
		# TYPE otf_runs_statuses gauge
`

	mc := &MetricsCollector{}
	mc.bootstrap()

	mc.update(pubsub.Event[*Event]{
		Payload: &Event{ID: testutils.ParseID(t, "run-1"), Status: runstatus.Pending},
	})
	mc.update(pubsub.Event[*Event]{
		Payload: &Event{ID: testutils.ParseID(t, "run-2"), Status: runstatus.Pending},
	})
	mc.update(pubsub.Event[*Event]{
		Payload: &Event{ID: testutils.ParseID(t, "run-2"), Status: runstatus.Planning},
	})
	mc.update(pubsub.Event[*Event]{
		Payload: &Event{ID: testutils.ParseID(t, "run-3"), Status: runstatus.Pending},
	})
	mc.update(pubsub.Event[*Event]{
		Payload: &Event{ID: testutils.ParseID(t, "run-3"), Status: runstatus.Planning},
	})
	mc.update(pubsub.Event[*Event]{
		Payload: &Event{ID: testutils.ParseID(t, "run-3"), Status: runstatus.Applied},
	})
	mc.update(pubsub.Event[*Event]{
		Payload: &Event{ID: testutils.ParseID(t, "run-4"), Status: runstatus.Pending},
	})
	mc.update(pubsub.Event[*Event]{
		Type:    pubsub.DeletedEvent,
		Payload: &Event{ID: testutils.ParseID(t, "run-4")},
	})

	assert.Len(t, mc.currentStatuses, 3)
	want := `
		otf_runs_statuses{status="applied"} 1
		otf_runs_statuses{status="pending"} 1
		otf_runs_statuses{status="planning"} 1
	`
	assert.NoError(t, testutil.CollectAndCompare(runStatusMetric, strings.NewReader(metadata+want), "otf_runs_statuses"))
}
