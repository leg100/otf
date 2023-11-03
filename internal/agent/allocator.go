package agent

import (
	"context"
	"fmt"
	"slices"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/workspace"
)

// allocator allocates jobs to agents. Only one allocator must be active on
// an OTF cluster.
type allocator struct {
	// Subscriber for receiving stream of job and agent events
	pubsub.Subscriber
	// service for seeding allocator with jobs and agents, and for
	// allocating jobs to agents.
	service
	// agents agents to allocate jobs to, keyed by agent ID
	agents map[string]*Agent
	// jobs jobs awaiting allocation to an agent, keyed by job ID
	jobs map[string]*Job
	// cache for looking up an agent's pool efficiently, keyed by pool ID
	pools map[string]*Pool
}

// Start the allocator. Should be invoked in a go routine.
func (a *allocator) Start(ctx context.Context) error {
	// Subscribe to job and agent events and unsubscribe before returning.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sub, err := a.Subscribe(ctx, "allocator-")
	if err != nil {
		return err
	}
	// seed allocator with pools, agents and jobs
	pools, err := a.listPools(ctx, listPoolOptions{})
	if err != nil {
		return err
	}
	agents, err := a.listAgents(ctx)
	if err != nil {
		return err
	}
	jobs, err := a.listJobs(ctx)
	if err != nil {
		return err
	}
	a.pools = make(map[string]*Pool, len(pools))
	for _, pool := range pools {
		a.pools[pool.ID] = pool
	}
	a.agents = make(map[string]*Agent, len(agents))
	for _, agent := range agents {
		a.agents[agent.ID] = agent
	}
	a.jobs = make(map[string]*Job, len(jobs))
	for _, job := range jobs {
		a.jobs[job.String()] = job
	}
	// now seeding has finished, allocate jobs
	a.allocate(ctx)
	// consume events until subscriber channel is closed.
	for event := range sub {
		switch payload := event.Payload.(type) {
		case *Pool:
			switch event.Type {
			case pubsub.DeletedEvent:
				delete(a.pools, payload.ID)
			default:
				a.pools[payload.ID] = payload
			}
		case *Agent:
			switch event.Type {
			case pubsub.DeletedEvent:
				delete(a.agents, payload.ID)
			default:
				a.agents[payload.ID] = payload
			}
		case *Job:
			switch event.Type {
			case pubsub.DeletedEvent:
				delete(a.jobs, payload.String())
			default:
				a.jobs[payload.String()] = payload
			}
		}
		a.allocate(ctx)
	}
	return pubsub.ErrSubscriptionTerminated
}

// allocate unallocated jobs to agents.
func (a *allocator) allocate(ctx context.Context) error {
	for _, job := range a.jobs {
		if job.Status != JobUnallocated {
			continue
		}
		for _, agent := range a.agents {
			if agent.Status != AgentIdle {
				// skip agents that are not idle and ready for jobs
				continue
			}
			switch job.ExecutionMode {
			case workspace.RemoteExecutionMode:
				// only server agents handle jobs with remote execution mode.
				if agent.Server {
					job.Status = JobAllocated
					agent.Concurrency--
					if err := a.allocateJob(ctx, job.RunID, agent.ID); err != nil {
						return err
					}
					return nil
				}
				continue
			case workspace.AgentExecutionMode:
				// only non-server agents handle jobs with agent execution mode.
				if agent.Server {
					continue
				}
				// non-server agents belong to a pool.
				pool, ok := a.pools[*agent.AgentPoolID]
				if !ok {
					return fmt.Errorf("missing cache entry for agent pool: %s", *agent.AgentPoolID)
				}
				if !slices.Contains(pool.Workspaces, job.WorkspaceID) {
					// job's workspace is configured to use a different pool
					continue
				}
				job.Status = JobAllocated
				agent.Concurrency--
				if err := a.allocateJob(ctx, job.RunID, agent.ID); err != nil {
					return err
				}
				return nil
			}
		}
	}
	return nil
}
