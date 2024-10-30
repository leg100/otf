package runner

import (
	"context"
	"fmt"
	"slices"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
)

// AllocatorLockID guarantees only one allocator on a cluster is running at any
// time.
const AllocatorLockID int64 = 5577006791947779412

// allocator allocates jobs to agents. Only one allocator must be active on
// an OTF cluster at any one time.
type allocator struct {
	logr.Logger
	// service for seeding and streaming pools, agents, and jobs, and for
	// allocating jobs to agents.
	client allocatorClient
	// cache of agent pools
	pools map[string]*Pool
	// runners to allocate jobs to, keyed by agent ID
	runners map[string]*runnerMeta
	// jobs awaiting allocation to an agent, keyed by job ID
	jobs map[JobSpec]*Job
}

type allocatorClient interface {
	WatchAgentPools(context.Context) (<-chan pubsub.Event[*Pool], func())
	WatchRunners(context.Context) (<-chan pubsub.Event[*runnerMeta], func())
	WatchJobs(context.Context) (<-chan pubsub.Event[*Job], func())

	listAllAgentPools(ctx context.Context) ([]*Pool, error)
	listRunners(ctx context.Context) ([]*runnerMeta, error)
	listJobs(ctx context.Context) ([]*Job, error)

	allocateJob(ctx context.Context, spec JobSpec, agentID string) (*Job, error)
	reallocateJob(ctx context.Context, spec JobSpec, agentID string) (*Job, error)
}

// Start the allocator. Should be invoked in a go routine.
func (a *allocator) Start(ctx context.Context) error {
	// Subscribe to pool, job and agent events and unsubscribe before returning.
	poolsSub, poolsUnsub := a.client.WatchAgentPools(ctx)
	defer poolsUnsub()
	agentsSub, agentsUnsub := a.client.WatchRunners(ctx)
	defer agentsUnsub()
	jobsSub, jobsUnsub := a.client.WatchJobs(ctx)
	defer jobsUnsub()

	// seed allocator with pools, agents, and jobs
	pools, err := a.client.listAllAgentPools(ctx)
	if err != nil {
		return err
	}
	agents, err := a.client.listRunners(ctx)
	if err != nil {
		return err
	}
	jobs, err := a.client.listJobs(ctx)
	if err != nil {
		return err
	}
	a.seed(pools, agents, jobs)

	// allocate jobs to agents
	a.allocate(ctx)

	// consume events until a subscriber is closed, and allocate jobs.
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
				delete(a.runners, event.Payload.ID)
			default:
				a.runners[event.Payload.ID] = event.Payload
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

func (a *allocator) seed(pools []*Pool, agents []*runnerMeta, jobs []*Job) {
	a.runners = make(map[string]*runnerMeta, len(agents))
	for _, agent := range agents {
		a.runners[agent.ID] = agent
	}
	a.jobs = make(map[JobSpec]*Job, len(jobs))
	for _, job := range jobs {
		a.jobs[job.Spec] = job
	}
}

// allocate jobs to runners.
func (a *allocator) allocate(ctx context.Context) error {
	for _, job := range a.jobs {
		var reallocate bool
		switch job.Status {
		case JobUnallocated:
			// see below
		case JobAllocated:
			// check agent the job is allocated to: if the agent is no longer in
			// a fit state then try to allocate job to another agent
			agent, ok := a.runners[*job.RunnerID]
			if !ok {
				return fmt.Errorf("agent %s not found in cache", *job.RunnerID)
			}
			if agent.Status == RunnerIdle || agent.Status == RunnerBusy {
				// agent still healthy, wait for agent to start job
				continue
			}
			// another no longer healthy, try reallocating job to another another
			reallocate = true
		case JobFinished, JobCanceled, JobErrored:
			// job has completed: remove and adjust number of current jobs
			// agents has
			delete(a.jobs, job.Spec)
			a.runners[*job.RunnerID].CurrentJobs--
			continue
		default:
			// job running; ignore
			continue
		}
		// allocate job to available agent
		var available []*runnerMeta
		for _, agent := range a.runners {
			if agent.Status != RunnerIdle && agent.Status != RunnerBusy {
				// skip agents that are not ready for jobs
				continue
			}
			// skip agents with insufficient capacity
			if agent.CurrentJobs == agent.MaxJobs {
				continue
			}
			if agent.AgentPoolID == nil {
				// if agent has a nil agent pool ID then it is a server
				// agent and it only handles jobs with a nil pool ID.
				if job.AgentPoolID != nil {
					continue
				}
			} else {
				// if an agent has a non-nil agent pool ID then it is a pool agent
				// and it only handles jobs with a matching pool ID.
				if job.AgentPoolID == nil || *agent.AgentPoolID != *job.AgentPoolID {
					continue
				}
			}
			available = append(available, agent)
		}
		if len(available) == 0 {
			a.Error(nil, "no available agents found for job", "job", job)
			continue
		}
		// select agent that has most recently sent a ping
		slices.SortFunc(available, func(a, b *runnerMeta) int {
			if a.LastPingAt.After(b.LastPingAt) {
				// a with more recent ping comes first in list
				return -1
			} else {
				return 1
			}
		})
		var (
			agent      = available[0]
			updatedJob *Job
			err        error
		)
		if reallocate {
			from := *job.RunnerID
			updatedJob, err = a.client.reallocateJob(ctx, job.Spec, agent.ID)
			if err != nil {
				return err
			}
			a.runners[from].CurrentJobs--
		} else {
			updatedJob, err = a.client.allocateJob(ctx, job.Spec, agent.ID)
			if err != nil {
				return err
			}
		}
		a.jobs[job.Spec] = updatedJob
		a.runners[agent.ID].CurrentJobs++
	}
	return nil
}
