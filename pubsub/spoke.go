package pubsub

import (
	"context"
	"fmt"
	"sync"

	"github.com/leg100/otf"
	"github.com/prometheus/client_golang/prometheus"
)

// subBufferSize is the buffer size of the channel for each subscription.
const subBufferSize = 16

var _ otf.PubSubService = (*spoke)(nil)

// spoke implements a 'pub-sub' service using go channels.
type spoke struct {
	mu      sync.Mutex
	subs    map[string]chan otf.Event
	metrics map[string]prometheus.Gauge
}

func newSpoke() *spoke {
	return &spoke{
		subs:    make(map[string]chan otf.Event),
		metrics: make(map[string]prometheus.Gauge),
	}
}

// Publish relays an event to a list of subscribers
func (e *spoke) Publish(event otf.Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for name, sub := range e.subs {
		// record sub's chan size
		e.metrics[name].Set(float64(len(sub)))

		// TODO: detect full channel using 'select...default:' and if full, close
		// the channel. Subs can re-subscribe if they wish (will have to
		// re-engineer subs first to handle this accordingly).
		sub <- event
	}
}

// Subscribe subscribes the caller to a stream of events.
func (e *spoke) Subscribe(ctx context.Context, name string) (<-chan otf.Event, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	sub := make(chan otf.Event, subBufferSize)
	if _, ok := e.subs[name]; ok {
		return nil, fmt.Errorf("name already taken")
	}
	e.subs[name] = sub

	totalSubscribers.Inc()

	e.metrics[name] = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "otf",
		Subsystem:   "pub_sub",
		Name:        "queue_length",
		Help:        "Total length for queue for subscriber",
		ConstLabels: prometheus.Labels{"name": name},
	})
	if err := prometheus.Register(e.metrics[name]); err != nil {
		return nil, err
	}

	// when the context is done remove the subscriber
	go func() {
		<-ctx.Done()

		totalSubscribers.Dec()

		e.mu.Lock()
		defer e.mu.Unlock()

		close(sub)
		delete(e.subs, name)

		prometheus.Unregister(e.metrics[name])
		delete(e.metrics, name)
	}()

	return sub, nil
}
