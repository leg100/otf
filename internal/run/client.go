package run

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	*otfapi.Client
}

func (c *Client) GetPlanFile(ctx context.Context, runID string, format PlanFormat) ([]byte, error) {
	u := fmt.Sprintf("runs/%s/planfile", url.QueryEscape(runID))
	req, err := c.NewRequest("GET", u, &PlanFileOptions{Format: format})
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	if err := c.Do(ctx, req, &buf); err != nil {
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
	opts := &PlanFileOptions{Format: format}
	q := url.Values{}
	if err := http.Encoder.Encode(opts, q); err != nil {
		return err
	}
	req.URL.RawQuery = q.Encode()

	if err := c.Do(ctx, req, nil); err != nil {
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
	if err := c.Do(ctx, req, &buf); err != nil {
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
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) ListRuns(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	req, err := c.NewRequest("GET", "runs", &opts)
	if err != nil {
		return nil, err
	}
	var list resource.Page[*Run]
	if err := c.Do(ctx, req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) Get(ctx context.Context, runID string) (*Run, error) {
	u := fmt.Sprintf("runs/%s", url.QueryEscape(runID))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	var run Run
	if err := c.Do(ctx, req, &run); err != nil {
		return nil, err
	}
	return &run, nil
}
