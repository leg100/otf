package http

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

func (c *Client) GetPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error) {
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

func (c *Client) UploadPlanFile(ctx context.Context, runID string, plan []byte, format otf.PlanFormat) error {
	u := fmt.Sprintf("runs/%s/planfile", url.QueryEscape(runID))
	req, err := c.NewRequest("PUT", u, plan)
	if err != nil {
		return err
	}

	// newRequest() only lets us set a query or a payload but not both, so we
	// set query here.
	opts := &planFileOptions{Format: format}
	q := url.Values{}
	if err := encoder.Encode(opts, q); err != nil {
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

func (c *Client) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	req, err := c.NewRequest("GET", "runs", &opts)
	if err != nil {
		return nil, err
	}

	wl := &jsonapi.RunList{}
	err = c.Do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalRunListJSONAPI(wl), nil
}

func (c *Client) GetRun(ctx context.Context, runID string) (*otf.Run, error) {
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

	return otf.UnmarshalRunJSONAPI(run), nil
}

func (c *Client) StartPhase(ctx context.Context, id string, phase otf.PhaseType, opts otf.PhaseStartOptions) (*otf.Run, error) {
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

	return otf.UnmarshalRunJSONAPI(run), nil
}

func (c *Client) FinishPhase(ctx context.Context, id string, phase otf.PhaseType, opts otf.PhaseFinishOptions) (*otf.Run, error) {
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

	return otf.UnmarshalRunJSONAPI(run), nil
}
