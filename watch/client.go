package watch

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	gohttp "net/http"
	"net/url"
	"path"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/run"
	"github.com/r3labs/sse/v2"
	"gopkg.in/cenkalti/backoff.v1"
)

type Client struct {
	http.Config
}

func (c *Client) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	// TODO: why buffered chan of size 1?
	notifications := make(chan otf.Event, 1)
	sseClient, err := newSSEClient(c.Config, notifications)
	if err != nil {
		return nil, err
	}

	go func() {
		err := sseClient.SubscribeRawWithContext(ctx, func(msg *sse.Event) {
			// TODO: impl support for objects other than runs
			if bytes.HasPrefix(msg.Event, []byte("run_")) {
				rawRun := jsonapi.Run{}
				if err := jsonapi.UnmarshalPayload(bytes.NewReader(msg.Data), &rawRun); err != nil {
					notifications <- otf.Event{Type: otf.EventError, Payload: err.Error()}
					return
				}
				notifications <- otf.Event{
					Type:    otf.EventType(msg.Event),
					Payload: run.NewFromJSONAPI(&rawRun),
				}
			}
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
	u.Path = path.Join(config.BasePath, config.WatchPath)

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
