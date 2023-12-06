package agent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/logr"
)

// rpcClient is a client for communication via RPC with the server.
type rpcClient struct {
	*otfapi.Client

	// agentID is the ID of the agent using the client
	agentID *string
	// address of OTF server
	address string
}

// NewRPCClient constructs a client that uses RPC to call OTF services. The
// agentID is added as an HTTP header to requests, allowing the server to
// identify the client; if nil then it is automatically added once the client
// successfully registers the agent.
func NewRPCClient(cfg otfapi.Config, agentID *string) (*rpcClient, error) {
	client, err := otfapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &rpcClient{
		Client:  client,
		address: cfg.Address,
	}, nil
}

func (c *rpcClient) NewJobClient(token []byte, logger logr.Logger) (*rpcClient, error) {
	return NewRPCClient(otfapi.Config{
		Address:       c.address,
		Token:         string(token),
		RetryRequests: true,
		RetryLogHook: func(_ retryablehttp.Logger, r *http.Request, n int) {
			// ignore first un-retried requests
			if n == 0 {
				return
			}
			logger.Error(nil, "retrying request", "url", r.URL, "attempt", n)
		},
	}, c.agentID)
}

// NewRequest constructs a new API request
func (c *rpcClient) NewRequest(method, path string, v interface{}) (*retryablehttp.Request, error) {
	req, err := c.Client.NewRequest(method, path, v)
	if err != nil {
		return nil, err
	}
	if c.agentID != nil {
		req.Header.Add(agentIDHeader, *c.agentID)
	}
	return req, err
}

func (c *rpcClient) registerAgent(ctx context.Context, opts registerAgentOptions) (*Agent, error) {
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

func (c *rpcClient) getAgentJobs(ctx context.Context, agentID string) ([]*Job, error) {
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

func (c *rpcClient) updateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error {
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

func (c *rpcClient) CreateAgentToken(ctx context.Context, poolID string, opts CreateAgentTokenOptions) (*agentToken, []byte, error) {
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

func (c *rpcClient) startJob(ctx context.Context, spec JobSpec) ([]byte, error) {
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

func (c *rpcClient) finishJob(ctx context.Context, spec JobSpec, opts finishJobOptions) error {
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
