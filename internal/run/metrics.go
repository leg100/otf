package run

import (
	"context"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(runStatusMetric)
}

var runStatusMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "runs",
	Name:      "statuses",
	Help:      "Total runs by status",
}, []string{"status"})

type MetricsCollector struct {
	service         metricsService
	currentStatuses map[resource.TfeID]runstatus.Status
}

type status struct {
	ID     resource.TfeID `db:"run_id"`
	Status runstatus.Status
}

type metricsService interface {
	Watch(ctx context.Context) (<-chan pubsub.Event[*Event], func())
	listStatuses(ctx context.Context) ([]status, error)
}

func (mc *MetricsCollector) Start(ctx context.Context) error {
	// subscribe to run events
	sub, unsub := mc.service.Watch(ctx)
	defer unsub()

	statuses, err := mc.service.listStatuses(ctx)
	if err != nil {
		return err
	}
	mc.bootstrap(statuses...)

	for event := range sub {
		mc.update(event)
	}
	return pubsub.ErrSubscriptionTerminated
}

func (mc *MetricsCollector) bootstrap(statuses ...status) {
	mc.currentStatuses = make(map[resource.TfeID]runstatus.Status, len(statuses))
	for _, run := range statuses {
		mc.currentStatuses[run.ID] = run.Status
		runStatusMetric.WithLabelValues(run.Status.String()).Inc()
	}
}

func (mc *MetricsCollector) update(event pubsub.Event[*Event]) {
	if event.Type == pubsub.DeletedEvent {
		// Run has been deleted, so lookup its last status and decrement
		// the tally.
		if lastStatus, ok := mc.currentStatuses[event.Payload.ID]; ok {
			runStatusMetric.WithLabelValues(lastStatus.String()).Dec()
			delete(mc.currentStatuses, event.Payload.ID)
		}
	} else {
		// Run has been created or updated.
		if lastStatus, ok := mc.currentStatuses[event.Payload.ID]; ok {
			// Decrement tally for its last status
			runStatusMetric.WithLabelValues(lastStatus.String()).Dec()
		}
		// Record new status
		mc.currentStatuses[event.Payload.ID] = event.Payload.Status
		// Increment tally for new status
		runStatusMetric.WithLabelValues(event.Payload.Status.String()).Inc()
	}
}
