package http

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

func (c *client) GetPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error) {
	u := fmt.Sprintf("runs/%s/planfile", url.QueryEscape(runID))
	req, err := c.newRequest("GET", u, &planFileOptions{Format: format})
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	err = c.do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *client) UploadPlanFile(ctx context.Context, runID string, plan []byte, format otf.PlanFormat) error {
	u := fmt.Sprintf("runs/%s/planfile", url.QueryEscape(runID))
	req, err := c.newRequest("PUT", u, plan)
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

	err = c.do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetLockFile(ctx context.Context, runID string) ([]byte, error) {
	u := fmt.Sprintf("runs/%s/lockfile", url.QueryEscape(runID))
	req, err := c.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	err = c.do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *client) UploadLockFile(ctx context.Context, runID string, lockfile []byte) error {
	u := fmt.Sprintf("runs/%s/lockfile", url.QueryEscape(runID))
	req, err := c.newRequest("PUT", u, lockfile)
	if err != nil {
		return err
	}

	err = c.do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	req, err := c.newRequest("GET", "runs", &opts)
	if err != nil {
		return nil, err
	}

	wl := &jsonapi.RunList{}
	err = c.do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalRunListJSONAPI(wl), nil
}

func (c *client) GetRun(ctx context.Context, runID string) (*otf.Run, error) {
	u := fmt.Sprintf("runs/%s", url.QueryEscape(runID))
	req, err := c.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	run := &jsonapi.Run{}
	err = c.do(ctx, req, run)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalRunJSONAPI(run), nil
}

func (c *client) StartPhase(ctx context.Context, id string, phase otf.PhaseType, opts otf.PhaseStartOptions) (*otf.Run, error) {
	u := fmt.Sprintf("runs/%s/actions/start/%s",
		url.QueryEscape(id),
		url.QueryEscape(string(phase)),
	)
	req, err := c.newRequest("POST", u, &opts)
	if err != nil {
		return nil, err
	}

	run := &jsonapi.Run{}
	err = c.do(ctx, req, run)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalRunJSONAPI(run), nil
}

func (c *client) FinishPhase(ctx context.Context, id string, phase otf.PhaseType, opts otf.PhaseFinishOptions) (*otf.Run, error) {
	u := fmt.Sprintf("runs/%s/actions/finish/%s",
		url.QueryEscape(id),
		url.QueryEscape(string(phase)),
	)
	req, err := c.newRequest("POST", u, &opts)
	if err != nil {
		return nil, err
	}

	run := &jsonapi.Run{}
	err = c.do(ctx, req, run)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalRunJSONAPI(run), nil
}
