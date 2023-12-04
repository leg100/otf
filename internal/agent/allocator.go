package agent

import (
	"context"
	"fmt"
	"slices"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/workspace"
)

// AllocatorLockID guarantees only one allocator on a cluster is running at any
// time.
const AllocatorLockID int64 = 5577006791947779412

// allocator allocates jobs to agents. Only one allocator must be active on
// an OTF cluster at any one time.
type allocator struct {
	// service for seeding and streaming pools, agents, and jobs, and for
	// allocating jobs to agents.
	Service
	// cache of agent pools
	pools map[string]*Pool
	// agents to allocate jobs to, keyed by agent ID
	agents map[string]*Agent
	// jobs awaiting allocation to an agent, keyed by job ID
	jobs map[JobSpec]*Job
}

// Start the allocator. Should be invoked in a go routine.
func (a *allocator) Start(ctx context.Context) error {
	// Subscribe to pool, job and agent events and unsubscribe before returning.
	poolsSub, poolsUnsub := a.watchAgentPools(ctx)
	defer poolsUnsub()
	agentsSub, agentsUnsub := a.watchAgents(ctx)
	defer agentsUnsub()
	jobsSub, jobsUnsub := a.watchJobs(ctx)
	defer jobsUnsub()

	// seed allocator with pools, agents, and jobs
	pools, err := a.listAllAgentPools(ctx)
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
	a.seed(pools, agents, jobs)

	// allocate jobs to agents
	a.allocate(ctx)

	// consume events until subscriber channel is closed, and allocate jobs.
	for {
		select {
		case event, open := <-poolsSub:
			if !open {
				return pubsub.ErrSubscriptionTerminated
			}
			switch event.Type {
			case pubsub.DeletedEvent:
				delete(a.pools, event.Payload.ID)
			default:
				a.pools[event.Payload.ID] = event.Payload
			}
		case event, open := <-agentsSub:
			if !open {
				return pubsub.ErrSubscriptionTerminated
			}
			switch event.Type {
			case pubsub.DeletedEvent:
				delete(a.agents, event.Payload.ID)
			default:
				a.agents[event.Payload.ID] = event.Payload
			}
		case event, open := <-jobsSub:
			if !open {
				return pubsub.ErrSubscriptionTerminated
			}
			switch event.Type {
			case pubsub.DeletedEvent:
				delete(a.jobs, event.Payload.Spec)
			default:
				a.jobs[event.Payload.Spec] = event.Payload
			}
		}
		if err := a.allocate(ctx); err != nil {
			return err
		}
	}
}

func (a *allocator) seed(pools []*Pool, agents []*Agent, jobs []*Job) {
	a.pools = make(map[string]*Pool, len(pools))
	for _, pool := range pools {
		a.pools[pool.ID] = pool
	}
	a.agents = make(map[string]*Agent, len(agents))
	for _, agent := range agents {
		a.agents[agent.ID] = agent
	}
	a.jobs = make(map[JobSpec]*Job, len(jobs))
	for _, job := range jobs {
		a.jobs[job.Spec] = job
	}
}

// allocate jobs to agents.
func (a *allocator) allocate(ctx context.Context) error {
	for _, job := range a.jobs {
		switch job.Status {
		case JobUnallocated:
			// allocate job to available agent
			if candidate, err := a.findCandidateAgent(job); err != nil {
				return err
			} else if candidate != nil {
				if err := a.allocateJob(ctx, candidate, job); err != nil {
					return err
				}
			}
		case JobAllocated:
			// check agent the job is allocated to, if the agent is no longer in a fit state then try to allocate job to another agent
			agent, ok := a.agents[*job.AgentID]
			if !ok {
				return fmt.Errorf("agent %s not found in cache", *job.AgentID)
			}
			if agent.Status == AgentIdle || agent.Status == AgentBusy {
				// agent still healthy, wait for agent to start job
				continue
			}
			// agent no longer healthy, try reallocating job to another agent
			if candidate, err := a.findCandidateAgent(job); err != nil {
				return err
			} else if candidate != nil {
				if err := a.reallocateJob(ctx, candidate, job); err != nil {
					return err
				}
			}
		case JobFinished, JobCanceled, JobErrored:
			// job has completed: remove and adjust number of current jobs
			// agents has
			delete(a.jobs, job.Spec)
			a.agents[*job.AgentID].CurrentJobs--
		}
	}
	return nil
}

// findCandidateAgent finds a suitable candidate agent for executing a job.
func (a *allocator) findCandidateAgent(job *Job) (*Agent, error) {
	var candidates []*Agent
	for _, agent := range a.agents {
		if agent.Status != AgentIdle && agent.Status != AgentBusy {
			// skip agents that are not ready for jobs
			continue
		}
		// skip agents with insufficient capacity
		if agent.CurrentJobs == agent.MaxJobs {
			continue
		}
		switch job.ExecutionMode {
		case workspace.RemoteExecutionMode:
			// only server agents handle jobs with remote execution mode.
			if agent.IsServer() {
				candidates = append(candidates, agent)
			}
			continue
		case workspace.AgentExecutionMode:
			// only pool agents handle jobs with agent execution mode.
			if agent.IsServer() {
				continue
			}
			// pool agents belong to a pool.
			pool, ok := a.pools[*agent.AgentPoolID]
			if !ok {
				return nil, fmt.Errorf("missing cache entry for agent pool: %s", *agent.AgentPoolID)
			}
			// pool can only handle jobs in same organization
			if pool.Organization != job.Organization {
				continue
			}
			if !pool.OrganizationScoped {
				if !slices.Contains(pool.AssignedWorkspaces, job.WorkspaceID) {
					// job's workspace is configured to use a different pool
					continue
				}
			}
			candidates = append(candidates, agent)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	// return agent that has most recently sent a ping
	slices.SortFunc(candidates, func(a, b *Agent) int {
		if a.LastPingAt.After(b.LastPingAt) {
			// a with more recent ping comes first in list
			return -1
		} else {
			return 1
		}
	})
	return candidates[0], nil
}

func (a *allocator) allocateJob(ctx context.Context, agent *Agent, job *Job) error {
	allocated, err := a.Service.allocateJob(ctx, job.Spec, agent.ID)
	if err != nil {
		return err
	}
	a.jobs[job.Spec] = allocated
	a.agents[agent.ID].CurrentJobs++
	return nil
}

func (a *allocator) reallocateJob(ctx context.Context, agent *Agent, job *Job) error {
	from := *job.AgentID
	reallocated, err := a.Service.reallocateJob(ctx, job.Spec, agent.ID)
	if err != nil {
		return err
	}
	a.jobs[job.Spec] = reallocated

	a.agents[from].CurrentJobs--
	a.agents[agent.ID].CurrentJobs++
	return nil
}
