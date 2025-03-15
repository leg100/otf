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

	listRunners(ctx context.Context, opts ListOptions) (*resource.Page[*RunnerMeta], error)
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

	runners, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*RunnerMeta], error) {
		return a.client.listRunners(ctx, ListOptions{PageOptions: opts})
	})
	if err != nil {
		return fmt.Errorf("seeding allocator with runners: %w", err)
	}
	jobs, err := a.client.listJobs(ctx)
	if err != nil {
		return fmt.Errorf("seeding allocator with jobs: %w", err)
	}
	a.seed(runners, jobs)

	// allocate jobs to runners
	for _, job := range a.jobs {
		if err := a.allocate(ctx, job); err != nil {
			return err
		}
	}

	// consume events until a subscriber is closed, and allocate jobs.
	for {
		select {
		case event, open := <-runnersSub:
			if !open {
				return pubsub.ErrSubscriptionTerminated
			}
			switch event.Type {
			case pubsub.CreatedEvent:
				a.addRunner(event.Payload)
			case pubsub.UpdatedEvent:
				switch event.Payload.Status {
				case RunnerExited, RunnerErrored:
					// Delete runners in terminal state.
					a.deleteRunner(event.Payload)
				default:
					a.runners[event.Payload.ID] = event.Payload
				}
			case pubsub.DeletedEvent:
				a.deleteRunner(event.Payload)
			}
		case event, open := <-jobsSub:
			if !open {
				return pubsub.ErrSubscriptionTerminated
			}
			switch event.Type {
			case pubsub.DeletedEvent:
				// Job is auto-deleted when its run is deleted (which occurs
				// when a workspace or org is deleted). If it was assigned to a
				// runner then decrement its current run tally.
				if job, ok := a.jobs[event.Payload.ID]; ok {
					if job.RunnerID != nil {
						a.decrementCurrentJobs(*job.RunnerID)
					}
				}
				delete(a.jobs, event.Payload.ID)
			default:
				a.jobs[event.Payload.ID] = event.Payload
			}
		}
		for _, job := range a.jobs {
			if err := a.allocate(ctx, job); err != nil {
				return err
			}
		}
	}
}

func (a *allocator) seed(runners []*RunnerMeta, jobs []*Job) {
	a.runners = make(map[resource.ID]*RunnerMeta, len(runners))
	a.currentJobs = make(map[resource.ID]int, len(runners))
	for _, runner := range runners {
		a.addRunner(runner)
	}
	a.jobs = make(map[resource.ID]*Job, len(jobs))
	for _, job := range jobs {
		// skip jobs in terminal state
		switch job.Status {
		case JobErrored, JobCanceled, JobFinished:
			continue
		}
		a.jobs[job.ID] = job
	}
}

// allocate jobs to runners.
func (a *allocator) allocate(ctx context.Context, job *Job) error {
	var reallocate bool
	switch job.Status {
	case JobAllocated:
		// check runner the job is allocated to: if the runner is no longer in
		// a fit state then try to allocate job to another runner
		runner, ok := a.runners[*job.RunnerID]
		if !ok {
			return fmt.Errorf("runner %s not found in cache", *job.RunnerID)
		}
		switch runner.Status {
		case RunnerIdle, RunnerBusy:
			// runner still healthy, wait for runner to start job
			return nil
		default:
			// no longer healthy, try reallocating job to another another
			a.Info("reallocating job away from unhealthy runner", "job", job, "runner", runner)
			reallocate = true
		}
	case JobFinished, JobCanceled, JobErrored:
		// job has completed
		delete(a.jobs, job.ID)
		// adjust current jobs of job's runner if allocated (an unallocated job
		// could have been canceled).
		if job.RunnerID != nil {
			a.decrementCurrentJobs(*job.RunnerID)
		}
		return nil
	case JobRunning:
		return nil
	case JobUnallocated:
		// proceed to allocate job below
	default:
		a.Error(nil, "unknown job status", "job", job)
		return nil
	}
	// allocate job to available runner
	var (
		available            []*RunnerMeta
		insufficientCapacity bool
	)
	for _, runner := range a.runners {
		// skip runners that are not ready for jobs
		if runner.Status != RunnerIdle && runner.Status != RunnerBusy {
			continue
		}
		// skip server runners for agent jobs
		if runner.AgentPool == nil && job.AgentPoolID != nil {
			continue
		}
		// skip agent runners for server jobs
		if runner.AgentPool != nil && job.AgentPoolID == nil {
			continue
		}
		// skip agent runners for agent jobs assigned to different agent
		// pool
		if runner.AgentPool != nil && job.AgentPoolID != nil && runner.AgentPool.ID != *job.AgentPoolID {
			continue
		}
		// skip runners with insufficient capacity
		if runner.MaxJobs == a.currentJobs[runner.ID] {
			insufficientCapacity = true
			continue
		}
		available = append(available, runner)
	}
	if len(available) == 0 {
		// If there is at least one appropriate runner but it has
		// insufficient capacity then it is a normal and temporary issue and
		// not worthy of reporting as an error.
		if !insufficientCapacity {
			a.Error(nil, "no available runners found for job", "job", job)
		}
		return nil
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
		a.decrementCurrentJobs(from)
	} else {
		updatedJob, err = a.client.allocateJob(ctx, job.ID, runner.ID)
		if err != nil {
			return err
		}
	}
	a.jobs[job.ID] = updatedJob
	a.incrementCurrentJobs(runner.ID)
	return nil
}

func (a *allocator) addRunner(runner *RunnerMeta) {
	// don't add runners in terminal state (exited, errored)
	switch runner.Status {
	case RunnerExited, RunnerErrored:
		return
	}
	a.runners[runner.ID] = runner
	a.currentJobs[runner.ID] = runner.CurrentJobs
	currentJobsMetric.WithLabelValues(runner.ID.String()).Set(float64(runner.CurrentJobs))
}

func (a *allocator) deleteRunner(runner *RunnerMeta) {
	delete(a.runners, runner.ID)
	delete(a.currentJobs, runner.ID)
	currentJobsMetric.DeleteLabelValues(runner.ID.String())
}

func (a *allocator) incrementCurrentJobs(runnerID resource.ID) {
	a.currentJobs[runnerID]++
	currentJobsMetric.WithLabelValues(runnerID.String()).Inc()
}

func (a *allocator) decrementCurrentJobs(runnerID resource.ID) {
	a.currentJobs[runnerID]--
	currentJobsMetric.WithLabelValues(runnerID.String()).Dec()
}
