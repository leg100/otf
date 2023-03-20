package watch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"net/url"
	"path"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/run"
	"github.com/r3labs/sse/v2"
	"gopkg.in/cenkalti/backoff.v1"
)

type Client struct {
	http.Config
}

// Watch returns a channel subscribed to events.
//
// NOTE: currently only subscribes to run events
func (c *Client) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	// TODO: why buffered chan of size 1?
	notifications := make(chan otf.Event, 1)
	sseClient, err := newSSEClient(c.Config, notifications)
	if err != nil {
		return nil, err
	}

	go func() {
		err := sseClient.SubscribeRawWithContext(ctx, func(raw *sse.Event) {
			event, err := unmarshal(raw)
			if err != nil {
				notifications <- otf.Event{Type: otf.EventError, Payload: err}
				return
			}
			notifications <- event
		})
		if err != nil {
			notifications <- otf.Event{Type: otf.EventError, Payload: err}
		}
		close(notifications)
	}()
	return notifications, nil
}

func newSSEClient(config http.Config, notifications chan otf.Event) (*sse.Client, error) {
	// construct watch URL endpoint
	addr, err := http.SanitizeAddress(config.Address)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %v", err)
	}
	u.Path = path.Join(config.BasePath, "/watch")

	client := sse.NewClient(u.String())
	client.EncodingBase64 = true
	// Disable backoff, it's instead the responsibility of the caller
	client.ReconnectStrategy = new(backoff.StopBackOff)
	client.OnConnect(func(_ *sse.Client) {
		notifications <- otf.Event{
			Type:    otf.EventInfo,
			Payload: "successfully connected",
		}
	})
	client.Headers = map[string]string{
		"Authorization": "Bearer " + config.Token,
	}
	if config.Insecure {
		client.Connection.Transport = &gohttp.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return client, nil
}

// unmarshal parses an SSE event and returns the equivalent OTF event
func unmarshal(event *sse.Event) (otf.Event, error) {
	if !strings.HasPrefix(string(event.Event), "run_") {
		return otf.Event{}, fmt.Errorf("no unmarshaler available for event %s", string(event.Event))
	}

	var run *run.Run
	if err := json.Unmarshal(event.Data, &run); err != nil {
		return otf.Event{}, err
	}

	return otf.Event{Type: otf.EventType(event.Event), Payload: run}, nil
}
