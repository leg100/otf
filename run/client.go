package run

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
	"github.com/r3labs/sse/v2"
	"gopkg.in/cenkalti/backoff.v1"
)

type Client struct {
	otf.JSONAPIClient
	http.Config
}

func (c *Client) GetPlanFile(ctx context.Context, runID string, format PlanFormat) ([]byte, error) {
	u := fmt.Sprintf("runs/%s/planfile", url.QueryEscape(runID))
	req, err := c.NewRequest("GET", u, &planFileOptions{Format: format})
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	err = c.Do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *Client) UploadPlanFile(ctx context.Context, runID string, plan []byte, format PlanFormat) error {
	u := fmt.Sprintf("runs/%s/planfile", url.QueryEscape(runID))
	req, err := c.NewRequest("PUT", u, plan)
	if err != nil {
		return err
	}

	// NewRequest() only lets us set a query or a payload but not both, so we
	// set query here.
	opts := &planFileOptions{Format: format}
	q := url.Values{}
	if err := http.Encoder.Encode(opts, q); err != nil {
		return err
	}
	req.URL.RawQuery = q.Encode()

	err = c.Do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetLockFile(ctx context.Context, runID string) ([]byte, error) {
	u := fmt.Sprintf("runs/%s/lockfile", url.QueryEscape(runID))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	err = c.Do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *Client) UploadLockFile(ctx context.Context, runID string, lockfile []byte) error {
	u := fmt.Sprintf("runs/%s/lockfile", url.QueryEscape(runID))
	req, err := c.NewRequest("PUT", u, lockfile)
	if err != nil {
		return err
	}

	err = c.Do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error) {
	req, err := c.NewRequest("GET", "runs", &opts)
	if err != nil {
		return nil, err
	}

	wl := &jsonapi.RunList{}
	err = c.Do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return newListFromJSONAPI(wl), nil
}

func (c *Client) GetRun(ctx context.Context, runID string) (*Run, error) {
	u := fmt.Sprintf("runs/%s", url.QueryEscape(runID))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	run := &jsonapi.Run{}
	err = c.Do(ctx, req, run)
	if err != nil {
		return nil, err
	}

	return newFromJSONAPI(run), nil
}

func (c *Client) StartPhase(ctx context.Context, id string, phase otf.PhaseType, opts PhaseStartOptions) (*Run, error) {
	u := fmt.Sprintf("runs/%s/actions/start/%s",
		url.QueryEscape(id),
		url.QueryEscape(string(phase)),
	)
	req, err := c.NewRequest("POST", u, &opts)
	if err != nil {
		return nil, err
	}

	run := &jsonapi.Run{}
	err = c.Do(ctx, req, run)
	if err != nil {
		return nil, err
	}

	return newFromJSONAPI(run), nil
}

func (c *Client) FinishPhase(ctx context.Context, id string, phase otf.PhaseType, opts PhaseFinishOptions) (*Run, error) {
	u := fmt.Sprintf("runs/%s/actions/finish/%s",
		url.QueryEscape(id),
		url.QueryEscape(string(phase)),
	)
	req, err := c.NewRequest("POST", u, &opts)
	if err != nil {
		return nil, err
	}

	run := &jsonapi.Run{}
	err = c.Do(ctx, req, run)
	if err != nil {
		return nil, err
	}

	return newFromJSONAPI(run), nil
}

// Watch returns a channel subscribed to run events.
func (c *Client) Watch(ctx context.Context, opts WatchOptions) (<-chan otf.Event, error) {
	// TODO: why buffered chan of size 1?
	notifications := make(chan otf.Event, 1)
	sseClient, err := newSSEClient(c.Config, notifications, opts)
	if err != nil {
		return nil, err
	}

	go func() {
		err := sseClient.SubscribeRawWithContext(ctx, func(raw *sse.Event) {
			run, err := UnmarshalJSONAPI(raw.Data)
			if err != nil {
				notifications <- otf.Event{Type: otf.EventError, Payload: err}
				return
			}
			notifications <- otf.Event{Type: otf.EventType(raw.Event), Payload: run}
		})
		if err != nil {
			notifications <- otf.Event{Type: otf.EventError, Payload: err}
		}
		close(notifications)
	}()
	return notifications, nil
}

func newSSEClient(config http.Config, notifications chan otf.Event, opts WatchOptions) (*sse.Client, error) {
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
	q := url.Values{}
	if err := http.Encoder.Encode(&opts, q); err != nil {
		return nil, err
	}
	u.RawQuery = q.Encode()

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
