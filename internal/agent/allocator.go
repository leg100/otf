package agent

import (
	"context"
	"fmt"
	"slices"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/workspace"
)

// allocator allocates jobs to agents. Only one allocator must be active on
// an OTF cluster at any one time.
type allocator struct {
	// Subscriber for receiving stream of job and agent events
	pubsub.Subscriber
	// service for seeding allocator with pools, agents, and jobs, and for
	// allocating jobs to agents.
	service
	// cache for looking up an agent's pool efficiently, keyed by pool ID
	pools map[string]*Pool
	// agents to allocate jobs to, keyed by agent ID
	agents map[string]*Agent
	// jobs awaiting allocation to an agent, keyed by job ID
	jobs map[JobSpec]*Job
	// capacities keeps track of the number of available workers each agent has,
	// keyed by agentID
	capacities map[string]int
}

// Start the allocator. Should be invoked in a go routine.
func (a *allocator) Start(ctx context.Context) error {
	// Subscribe to job and agent events and unsubscribe before returning.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sub, err := a.Subscribe(ctx, "job-allocator-")
	if err != nil {
		return err
	}
	// seed allocator with pools, agents, capacities, and jobs
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
	a.capacities = make(map[string]int, len(agents))
	for _, agent := range agents {
		a.agents[agent.ID] = agent
		a.capacities[agent.ID] = agent.Concurrency
	}
	a.jobs = make(map[JobSpec]*Job, len(jobs))
	for _, job := range jobs {
		a.jobs[job.JobSpec] = job
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
				delete(a.capacities, payload.ID)
			default:
				if _, ok := a.agents[payload.ID]; !ok {
					// new agent, initialize its capacity
					a.capacities[payload.ID] = payload.Concurrency
				}
				a.agents[payload.ID] = payload
			}
		case *Job:
			switch event.Type {
			case pubsub.DeletedEvent:
				delete(a.jobs, payload.JobSpec)
			default:
				a.jobs[payload.JobSpec] = payload
			}
		}
		a.allocate(ctx)
	}
	return pubsub.ErrSubscriptionTerminated
}

// allocate jobs to agents.
func (a *allocator) allocate(ctx context.Context) error {
	allocatefn := func(agent *Agent, job *Job) error {
		job.Status = JobAllocated
		job.AgentID = agent.ID
		a.capacities[agent.ID]--
		return a.allocateJob(ctx, job)
	}
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
					return allocatefn(agent, job)
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
				return allocatefn(agent, job)
			}
		}
	}
	return nil
}
