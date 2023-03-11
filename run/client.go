package run

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
)

type Client struct {
	otf.JSONAPIClient
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

func (c *Client) ListRuns(ctx context.Context, opts *RunListOptions) (*RunList, error) {
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
