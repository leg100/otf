package http

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
)

func (c *client) GetPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error) {
	u := fmt.Sprintf("runs/%s/planfile", url.QueryEscape(runID))
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

func (c *client) UploadPlanFile(ctx context.Context, runID string, plan []byte, format otf.PlanFormat) error {
	u := fmt.Sprintf("runs/%s/planfile", url.QueryEscape(runID))
	req, err := c.newRequest("PUT", u, plan)
	if err != nil {
		return err
	}

	// newRequest() only lets us set a query or a payload but not both, so we
	// set query here.
	opts := &uploadPlanFileOptions{Format: format}
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
