package runner

import (
	"bytes"
	"context"
	"fmt"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/resource"

	"github.com/hashicorp/go-retryablehttp"
)

const runnerIDHeaderKey = "otf-agent-id"

type client interface {
	register(ctx context.Context, opts registerOptions) (*RunnerMeta, error)
	updateStatus(ctx context.Context, agentID resource.TfeID, status RunnerStatus) error

	awaitAllocatedJobs(ctx context.Context, agentID resource.TfeID) ([]*Job, error)
	startJob(ctx context.Context, jobID resource.TfeID) ([]byte, error)
	finishJob(ctx context.Context, jobID resource.TfeID, opts finishJobOptions) error
}

// client accesses the service endpoints via RPC.
type remoteClient struct {
	*otfapi.Client

	// agentID is the ID of the agent using the client
	agentID *resource.TfeID
}

// newRequest constructs a new API request
func (c *remoteClient) newRequest(method, path string, v interface{}) (*retryablehttp.Request, error) {
	req, err := c.Client.NewRequest(method, path, v)
	if err != nil {
		return nil, err
	}
	if c.agentID != nil {
		req.Header.Add(runnerIDHeaderKey, c.agentID.String())
	}
	return req, err
}

func (c *remoteClient) register(ctx context.Context, opts registerOptions) (*RunnerMeta, error) {
	req, err := c.newRequest("POST", "agents/register", &opts)
	if err != nil {
		return nil, err
	}
	var m RunnerMeta
	if err := c.Do(ctx, req, &m); err != nil {
		return nil, err
	}
	// add agent ID to future requests
	agentID := m.ID
	c.agentID = &agentID
	return &m, nil
}

func (c *remoteClient) awaitAllocatedJobs(ctx context.Context, agentID resource.TfeID) ([]*Job, error) {
	req, err := c.newRequest("GET", "agents/jobs", nil)
	if err != nil {
		return nil, err
	}

	var jobs []*Job
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

func (c *remoteClient) awaitJobSignal(ctx context.Context, jobID resource.TfeID) func() (jobSignal, error) {
	u := fmt.Sprintf("jobs/%s/await-signal", jobID)
	req, err := c.newRequest("GET", u, nil)
	if err != nil {
		return func() (jobSignal, error) {
			return jobSignal{}, nil
		}
	}

	var signal jobSignal
	// GET request blocks until:
	// (a) job signal is sent
	// (b) a timeout is reached
	//
	// (b) can occur due to any intermediate proxies placed between the client
	// and otfd, such as nginx, which has a default proxy_read_timeout of 60s.
	if err := c.Do(ctx, req, &signal); err != nil {
		return func() (jobSignal, error) {
			return jobSignal{}, err
		}
	}
	return func() (jobSignal, error) {
		return signal, nil
	}
}

func (c *remoteClient) updateStatus(ctx context.Context, agentID resource.TfeID, status RunnerStatus) error {
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

func (c *remoteClient) CreateAgentToken(ctx context.Context, poolID resource.TfeID, opts CreateAgentTokenOptions) (*agentToken, []byte, error) {
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

func (c *remoteClient) startJob(ctx context.Context, jobID resource.TfeID) ([]byte, error) {
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

func (c *remoteClient) finishJob(ctx context.Context, jobID resource.TfeID, opts finishJobOptions) error {
	req, err := c.newRequest("POST", "jobs/finish", &finishJobParams{
		JobID:            jobID,
		finishJobOptions: opts,
	})
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

func (c *remoteClient) GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error) {
	u := fmt.Sprintf("jobs/%s/dynamic-credentials", jobID)
	req, err := c.newRequest("POST", u, generateDynamicCredentialsTokenParams{
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
