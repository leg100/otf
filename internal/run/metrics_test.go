package run

import (
	"strings"
	"testing"

	"github.com/leg100/otf/internal/pubsub"
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
		&Run{ID: "run-1", Status: RunPending},
		&Run{ID: "run-2", Status: RunPending},
		&Run{ID: "run-3", Status: RunPending},
		&Run{ID: "run-4", Status: RunPending},
		&Run{ID: "run-5", Status: RunPlanning},
		&Run{ID: "run-6", Status: RunPlanning},
		&Run{ID: "run-7", Status: RunPlanning},
		&Run{ID: "run-8", Status: RunPlanning},
		&Run{ID: "run-9", Status: RunApplied},
		&Run{ID: "run-10", Status: RunApplied},
		&Run{ID: "run-11", Status: RunApplied},
		&Run{ID: "run-12", Status: RunApplied},
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

	mc.update(pubsub.Event[*Run]{
		Payload: &Run{ID: "run-1", Status: RunPending},
	})
	mc.update(pubsub.Event[*Run]{
		Payload: &Run{ID: "run-2", Status: RunPending},
	})
	mc.update(pubsub.Event[*Run]{
		Payload: &Run{ID: "run-2", Status: RunPlanning},
	})
	mc.update(pubsub.Event[*Run]{
		Payload: &Run{ID: "run-3", Status: RunPending},
	})
	mc.update(pubsub.Event[*Run]{
		Payload: &Run{ID: "run-3", Status: RunPlanning},
	})
	mc.update(pubsub.Event[*Run]{
		Payload: &Run{ID: "run-3", Status: RunApplied},
	})
	mc.update(pubsub.Event[*Run]{
		Payload: &Run{ID: "run-4", Status: RunPending},
	})
	mc.update(pubsub.Event[*Run]{
		Type:    pubsub.DeletedEvent,
		Payload: &Run{ID: "run-4"},
	})

	assert.Len(t, mc.currentStatuses, 3)
	want := `

		otf_runs_statuses{status="applied"} 1
		otf_runs_statuses{status="pending"} 1
		otf_runs_statuses{status="planning"} 1
	`
	assert.NoError(t, testutil.CollectAndCompare(runStatusMetric, strings.NewReader(metadata+want), "otf_runs_statuses"))
}
