package api

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runner"
)

// Alias client to permit embedding it with other clients in a struct
// without a name clash.
type RunnerClient = Client

// client accesses the service endpoints via RPC.
type Client struct {
	*otfhttp.Client

	// agentID is the ID of the agent using the client
	agentID *resource.TfeID
}

// newRequest constructs a new API request
func (c *Client) newRequest(method, path string, v any) (*retryablehttp.Request, error) {
	req, err := c.Client.NewRequest(method, path, v)
	if err != nil {
		return nil, err
	}
	if c.agentID != nil {
		req.Header.Add(runner.RunnerIDHeaderKey, c.agentID.String())
	}
	return req, err
}

func (c *Client) Register(ctx context.Context, opts runner.RegisterRunnerOptions) (*runner.RunnerMeta, error) {
	req, err := c.newRequest("POST", "agents/register", &opts)
	if err != nil {
		return nil, err
	}
	var m runner.RunnerMeta
	if err := c.Do(ctx, req, &m); err != nil {
		return nil, err
	}
	// add agent ID to future requests
	c.agentID = &m.ID
	return &m, nil
}

func (c *Client) AwaitAllocatedJobs(ctx context.Context, agentID resource.TfeID) ([]*runner.Job, error) {
	req, err := c.newRequest("GET", "agents/await-allocated-jobs", nil)
	if err != nil {
		return nil, err
	}

	var jobs []*runner.Job
	// GET request blocks until:
	//
	// (a) job(s) are allocated to agent
	// (b) a timeout is reached
	//
	// (b) can occur due to any intermediate proxies placed between otf-agent
	// and otfd, such as nginx, which has a default proxy_read_timeout of 60s.
	if err := c.Do(ctx, req, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *Client) GetJob(ctx context.Context, jobID resource.TfeID) (*runner.Job, error) {
	u := fmt.Sprintf("jobs/%s", jobID)
	req, err := c.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	var job runner.Job
	if err := c.Do(ctx, req, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

//lint:ignore U1000 staticcheck wrongly picks this up as unused - perhaps because only the RunnerClient alias is used?
func (c *Client) awaitJobSignal(ctx context.Context, jobID resource.TfeID) func() (runner.JobSignal, error) {
	u := fmt.Sprintf("jobs/%s/await-signal", jobID)
	req, err := c.newRequest("GET", u, nil)
	if err != nil {
		return func() (runner.JobSignal, error) {
			return runner.JobSignal{}, nil
		}
	}

	var signal runner.JobSignal
	// GET request blocks until:
	// (a) job signal is sent
	// (b) a timeout is reached
	//
	// (b) can occur due to any intermediate proxies placed between the client
	// and otfd, such as nginx, which has a default proxy_read_timeout of 60s.
	if err := c.Do(ctx, req, &signal); err != nil {
		return func() (runner.JobSignal, error) {
			return runner.JobSignal{}, err
		}
	}
	return func() (runner.JobSignal, error) {
		return signal, nil
	}
}

func (c *Client) UpdateStatus(ctx context.Context, agentID resource.TfeID, status runner.RunnerStatus) error {
	req, err := c.newRequest("POST", "agents/status", &updateAgentStatusParams{
		Status: status,
	})
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

// agent tokens

func (c *Client) CreateAgentToken(ctx context.Context, poolID resource.TfeID, opts runner.CreateAgentTokenOptions) (*runner.AgentToken, []byte, error) {
	u := fmt.Sprintf("agent-tokens/%s/create", poolID)
	req, err := c.newRequest("POST", u, &opts)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, nil, err
	}
	return nil, buf.Bytes(), nil
}

// jobs

func (c *Client) StartJob(ctx context.Context, jobID resource.TfeID) ([]byte, error) {
	req, err := c.newRequest("POST", "jobs/start", &startJobParams{
		JobID: jobID,
	})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//lint:ignore U1000 staticcheck wrongly picks this up as unused - perhaps because only the RunnerClient alias is used?
func (c *Client) finishJob(ctx context.Context, jobID resource.TfeID, opts runner.FinishJobOptions) error {
	req, err := c.newRequest("POST", "jobs/finish", &finishJobParams{
		JobID:            jobID,
		FinishJobOptions: opts,
	})
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error) {
	u := fmt.Sprintf("jobs/%s/dynamic-credentials", jobID)
	req, err := c.newRequest("POST", u, &generateDynamicCredentialsTokenParams{
		Audience: audience,
	})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
