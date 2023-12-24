package agent

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
	otfapi "github.com/leg100/otf/internal/api"
)

const agentIDHeader = "otf-agent-id"

// client accesses the service endpoints via RPC.
type client struct {
	*otfapi.Client

	// agentID is the ID of the agent using the client
	agentID *string
}

// NewRequest constructs a new API request
func (c *client) NewRequest(method, path string, v interface{}) (*retryablehttp.Request, error) {
	req, err := c.Client.NewRequest(method, path, v)
	if err != nil {
		return nil, err
	}
	if c.agentID != nil {
		req.Header.Add(agentIDHeader, *c.agentID)
	}
	return req, err
}

func (c *client) registerAgent(ctx context.Context, opts registerAgentOptions) (*Agent, error) {
	req, err := c.NewRequest("POST", "agents/register", &opts)
	if err != nil {
		return nil, err
	}
	var agent Agent
	if err := c.Do(ctx, req, &agent); err != nil {
		return nil, err
	}
	// add agent ID to future requests
	agentID := agent.ID
	c.agentID = &agentID
	return &agent, nil
}

func (c *client) getAgentJobs(ctx context.Context, agentID string) ([]*Job, error) {
	req, err := c.NewRequest("GET", "agents/jobs", nil)
	if err != nil {
		return nil, err
	}

	var jobs []*Job
	// GET request blocks until:
	//
	// (a) job(s) are allocated to agent
	// (b) job(s) already allocated to agent are sent a cancelation signal
	// (c) a timeout is reached
	//
	// (c) can occur due to any intermediate proxies placed between otf-agent
	// and otfd, such as nginx, which has a default proxy_read_timeout of 60s.
	if err := c.Do(ctx, req, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *client) updateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error {
	req, err := c.NewRequest("POST", "agents/status", &updateAgentStatusParams{
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

func (c *client) CreateAgentToken(ctx context.Context, poolID string, opts CreateAgentTokenOptions) (*agentToken, []byte, error) {
	u := fmt.Sprintf("agent-tokens/%s/create", poolID)
	req, err := c.NewRequest("POST", u, &opts)
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

func (c *client) startJob(ctx context.Context, spec JobSpec) ([]byte, error) {
	req, err := c.NewRequest("POST", "agents/start", &spec)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *client) finishJob(ctx context.Context, spec JobSpec, opts finishJobOptions) error {
	req, err := c.NewRequest("POST", "agents/finish", &finishJobParams{
		JobSpec:          spec,
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

// agent pools
func (c *client) CreateAgentPool(ctx context.Context, opts CreateAgentPoolOptions) (*Pool, error) {
	req, err := c.NewRequest("POST", "agent-pools/create", &opts)
	if err != nil {
		return nil, err
	}
	var pool Pool
	if err := c.Do(ctx, req, &pool); err != nil {
		return nil, err
	}
	return &pool, nil
}
