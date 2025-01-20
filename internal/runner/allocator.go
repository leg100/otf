package runner

import (
	"context"
	"fmt"
	"slices"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
)

// allocator allocates jobs to runners. Only one allocator must be active on
// an OTF cluster at any one time.
type allocator struct {
	logr.Logger
	// service for seeding and streaming pools, runners, and jobs, and for
	// allocating jobs to runners.
	client allocatorClient
	// runners keyed by runner ID
	runners map[resource.ID]*RunnerMeta
	// jobs keyed by job ID
	jobs map[resource.ID]*Job
	// total current jobs allocated to each runner keyed by runner ID
	currentJobs map[resource.ID]int
}

type allocatorClient interface {
	WatchRunners(context.Context) (<-chan pubsub.Event[*RunnerMeta], func())
	WatchJobs(context.Context) (<-chan pubsub.Event[*Job], func())

	listRunners(ctx context.Context) ([]*RunnerMeta, error)
	listJobs(ctx context.Context) ([]*Job, error)

	allocateJob(ctx context.Context, jobID, runnerID resource.ID) (*Job, error)
	reallocateJob(ctx context.Context, jobID, runnerID resource.ID) (*Job, error)
}

// Start the allocator. Should be invoked in a go routine.
func (a *allocator) Start(ctx context.Context) error {
	// Subscribe to pool, job and runner events and unsubscribe before returning.
	runnersSub, runnersUnsub := a.client.WatchRunners(ctx)
	defer runnersUnsub()
	jobsSub, jobsUnsub := a.client.WatchJobs(ctx)
	defer jobsUnsub()

	runners, err := a.client.listRunners(ctx)
	if err != nil {
		return fmt.Errorf("seeding allocator with runners: %w", err)
	}
	jobs, err := a.client.listJobs(ctx)
	if err != nil {
		return fmt.Errorf("seeding allocator with jobs: %w", err)
	}
	a.seed(runners, jobs)

	// allocate jobs to runners
	a.allocate(ctx)

	// consume events until a subscriber is closed, and allocate jobs.
	for {
		select {
		case event, open := <-runnersSub:
			if !open {
				return pubsub.ErrSubscriptionTerminated
			}
			switch event.Type {
			case pubsub.DeletedEvent:
				delete(a.runners, event.Payload.ID)
				delete(a.currentJobs, event.Payload.ID)
			default:
				a.runners[event.Payload.ID] = event.Payload
			}
		case event, open := <-jobsSub:
			if !open {
				return pubsub.ErrSubscriptionTerminated
			}
			switch event.Type {
			case pubsub.DeletedEvent:
				delete(a.jobs, event.Payload.ID)
			default:
				a.jobs[event.Payload.ID] = event.Payload
			}
		}
		if err := a.allocate(ctx); err != nil {
			return err
		}
	}
}

func (a *allocator) seed(runners []*RunnerMeta, jobs []*Job) {
	a.runners = make(map[resource.ID]*RunnerMeta, len(runners))
	a.currentJobs = make(map[resource.ID]int, len(runners))
	for _, runner := range runners {
		a.runners[runner.ID] = runner
		a.currentJobs[runner.ID] = runner.CurrentJobs
	}
	a.jobs = make(map[resource.ID]*Job, len(jobs))
	for _, job := range jobs {
		a.jobs[job.ID] = job
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
			// check runner the job is allocated to: if the runner is no longer in
			// a fit state then try to allocate job to another runner
			runner, ok := a.runners[*job.RunnerID]
			if !ok {
				return fmt.Errorf("runner %s not found in cache", *job.RunnerID)
			}
			if runner.Status == RunnerIdle || runner.Status == RunnerBusy {
				// runner still healthy, wait for runner to start job
				continue
			}
			// no longer healthy, try reallocating job to another another
			a.Info("reallocating job away from unhealthy runner", "job", job, "runner", runner)
			reallocate = true
		case JobFinished, JobCanceled, JobErrored:
			// job has completed: remove and adjust number of current jobs
			// runner has
			delete(a.jobs, job.ID)
			a.currentJobs[*job.RunnerID]--
			continue
		default:
			// job running; ignore
			continue
		}
		// allocate job to available runner
		var available []*RunnerMeta
		for _, runner := range a.runners {
			if runner.Status != RunnerIdle && runner.Status != RunnerBusy {
				// skip runners that are not ready for jobs
				continue
			}
			// skip runners with insufficient capacity
			if a.currentJobs[runner.ID] == runner.MaxJobs {
				continue
			}
			if runner.AgentPool == nil {
				// if runner has a nil agent pool ID then it is a server
				// runner and it only handles jobs with a nil pool ID.
				if job.AgentPoolID != nil {
					continue
				}
			} else {
				// if a runner has a non-nil agent pool ID then it is an agent
				// and it only handles jobs with a matching pool ID.
				if job.AgentPoolID == nil || runner.AgentPool.ID != *job.AgentPoolID {
					continue
				}
			}
			available = append(available, runner)
		}
		if len(available) == 0 {
			a.Error(nil, "no available runners found for job", "job", job)
			continue
		}
		// select runner that has most recently sent a ping
		slices.SortFunc(available, func(a, b *RunnerMeta) int {
			if a.LastPingAt.After(b.LastPingAt) {
				// a with more recent ping comes first in list
				return -1
			} else {
				return 1
			}
		})
		var (
			runner     = available[0]
			updatedJob *Job
			err        error
		)
		if reallocate {
			from := *job.RunnerID
			updatedJob, err = a.client.reallocateJob(ctx, job.ID, runner.ID)
			if err != nil {
				return err
			}
			a.currentJobs[from]--
		} else {
			updatedJob, err = a.client.allocateJob(ctx, job.ID, runner.ID)
			if err != nil {
				return err
			}
		}
		a.jobs[job.ID] = updatedJob
		a.currentJobs[runner.ID]++
	}
	return nil
}
