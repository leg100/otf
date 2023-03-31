package pubsub

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/leg100/otf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroker_Subscribe(t *testing.T) {
	broker := &Broker{
		subs:    make(map[string]chan otf.Event),
		metrics: make(map[string]prometheus.Gauge),
	}

	ctx, cancel := context.WithCancel(context.Background())

	sub, err := broker.Subscribe(ctx, "")
	require.NoError(t, err)

	assert.Equal(t, 1, len(broker.subs))

	cancel()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_Publish(t *testing.T) {
	got := make(chan otf.Event, 1)
	pool := &fakePool{}
	broker := &Broker{
		pool:          pool,
		subs:          map[string]chan otf.Event{"sub-1": got},
		registrations: make(map[string]otf.Getter),
		metrics:       map[string]prometheus.Gauge{"sub-1": prometheus.NewGauge(prometheus.GaugeOpts{})},
	}

	type payload struct {
		ID string
	}

	event := otf.Event{
		Type:    otf.EventType("payload_update"),
		Payload: &payload{ID: "payload-123"},
	}
	broker.Publish(event)

	// locally published event
	assert.Equal(t, event, <-broker.subs["sub-1"])

	// remotely published message
	if assert.Equal(t, 1, len(pool.gotExecArgs)) {
		var msg pgevent
		err := json.Unmarshal(pool.gotExecArgs[0].([]byte), &msg)
		require.NoError(t, err)
		want := pgevent{PayloadType: "*pubsub.payload", Event: "payload_update", ID: "payload-123"}
		assert.Equal(t, want, msg)
	}
}

func TestPubSub_receive(t *testing.T) {
	notification := pgconn.Notification{
		Payload: "{\"payload_type\":\"run\",\"event\":\"run_status_update\",\"id\":\"run-123\",\"pid\":\"process-1\"}",
	}
	resource := struct {
		ID string
	}{
		ID: "run-123",
	}
	got := make(chan otf.Event, 1)
	broker := &Broker{
		pool:          &fakePool{},
		subs:          map[string]chan otf.Event{"sub-1": got},
		registrations: map[string]otf.Getter{"run": &fakeGetter{resource: resource}},
		metrics:       map[string]prometheus.Gauge{"sub-1": prometheus.NewGauge(prometheus.GaugeOpts{})},
	}
	err := broker.receive(context.Background(), &notification)
	require.NoError(t, err)

	want := otf.Event{
		Type:    otf.EventRunStatusUpdate,
		Payload: resource,
	}
	assert.Equal(t, want, <-got)
}
