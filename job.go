package otf

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgtype"
)

const (
	JobCanceled    JobStatus = "canceled"
	JobErrored     JobStatus = "errored"
	JobPending     JobStatus = "pending"
	JobQueued      JobStatus = "queued"
	JobRunning     JobStatus = "running"
	JobFinished    JobStatus = "finished"
	JobUnreachable JobStatus = "unreachable"
)

var (
	ErrJobAlreadyClaimed = errors.New("job already claimed")
)

// Job is a unit of work
type Job interface {
	// Do some work in an execution environment
	Do(Environment) error
	// ID identifies the work
	JobID() string
	// Status provides the current status of the job
	Status() JobStatus
}

type JobStatus string

// job is functionality common to all jobs
type job struct {
	id                     string
	status                 JobStatus
	statusTimestamps       []JobStatusTimestamp
	runID                  string
	configurationVersionID string
	workspaceID            string
	// terraform flags
	isDestroy bool
}

func (j *job) JobID() string                          { return j.id }
func (j *job) Status() JobStatus                      { return j.status }
func (j *job) StatusTimestamps() []JobStatusTimestamp { return j.statusTimestamps }

// StatusTimestamp retrieves the timestamp for a status.
func (j *job) StatusTimestamp(status JobStatus) (time.Time, error) {
	for _, rst := range j.statusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, ErrStatusTimestampNotFound
}

func (j *job) UpdateStatus(status JobStatus) {
	j.status = status
	j.statusTimestamps = append(j.statusTimestamps, JobStatusTimestamp{
		Status:    status,
		Timestamp: CurrentTimestamp(),
	})
}

// setupEnv invokes the necessary steps before a plan or apply can proceed.
func (j *job) setup(env Environment) error {
	if err := env.RunFunc(j.downloadConfig); err != nil {
		return err
	}
	err := env.RunFunc(func(ctx context.Context, env Environment) error {
		return deleteBackendConfigFromDirectory(ctx, env.Path())
	})
	if err != nil {
		return err
	}
	if err := env.RunFunc(j.downloadState); err != nil {
		return err
	}
	if err := env.RunCLI("terraform", "init"); err != nil {
		return fmt.Errorf("running terraform init: %w", err)
	}
	return nil
}

func (j *job) downloadConfig(ctx context.Context, env Environment) error {
	// Download config
	cv, err := env.ConfigurationVersionService().Download(ctx, j.configurationVersionID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}
	// Decompress and untar config
	if err := Unpack(bytes.NewBuffer(cv), env.Path()); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}
	return nil
}

// downloadState downloads current state to disk. If there is no state yet
// nothing will be downloaded and no error will be reported.
func (j *job) downloadState(ctx context.Context, env Environment) error {
	state, err := env.StateVersionService().Current(ctx, j.workspaceID)
	if errors.Is(err, ErrResourceNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("retrieving current state version: %w", err)
	}
	statefile, err := env.StateVersionService().Download(ctx, state.ID())
	if err != nil {
		return fmt.Errorf("downloading state version: %w", err)
	}
	if err := os.WriteFile(filepath.Join(env.Path(), LocalStateFilename), statefile, 0644); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}
	return nil
}

func (j *job) downloadPlanFile(ctx context.Context, env Environment) error {
	plan, err := env.RunService().GetPlanFile(ctx, RunGetOptions{ID: String(j.runID)}, PlanFormatBinary)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(env.Path(), PlanFilename), plan, 0644)
}

// uploadState reads, parses, and uploads terraform state
func (j *job) uploadState(ctx context.Context, env Environment) error {
	f, err := os.ReadFile(filepath.Join(env.Path(), LocalStateFilename))
	if err != nil {
		return err
	}
	state, err := UnmarshalState(f)
	if err != nil {
		return err
	}
	_, err = env.StateVersionService().Create(ctx, j.workspaceID, StateVersionCreateOptions{
		State:   String(base64.StdEncoding.EncodeToString(f)),
		MD5:     String(fmt.Sprintf("%x", md5.Sum(f))),
		Lineage: &state.Lineage,
		Serial:  Int64(state.Serial),
		RunID:   &j.runID,
	})
	if err != nil {
		return err
	}
	return nil
}

func newJob() *job {
	return &job{
		id:     NewID("job"),
		status: JobPending,
	}
}

type JobService interface {
	// Queued returns a list of queued jobs
	Queued(ctx context.Context) ([]*Job, error)
	// Claim claims a job entitling the caller to do the job.
	// ErrJobAlreadyClaimed is returned if job is already claimed.
	Claim(ctx context.Context, id string, opts JobClaimOptions) (*Job, error)
	// Finish is called by an agent when it finishes a job.
	Finish(ctx context.Context, id string, opts JobFinishOptions) (*Job, error)
	// PutChunk uploads a chunk of logs from the job.
	PutChunk(ctx context.Context, id string, chunk Chunk) error
}

type JobClaimOptions struct {
	AgentID string
}

type JobFinishOptions struct {
	Errored bool
}

// JobStore persists jobs
type JobStore interface {
	Create(ctx context.Context, job *Job) error
}

type JobStatusTimestamp struct {
	Status    JobStatus
	Timestamp time.Time
}

type PlanJob struct {
	// for retrieving config tarball
	configurationVersionID string
	// for retrieving latest state
	workspaceID string
	// for uploading plan file
	planID string
	// flags for terraform plan
	isDestroy bool
}

type ApplyJob struct {
	// for retrieving config tarball
	configurationVersionID string
	// for retrieving latest state and creating new state
	workspaceID string
	// for retrieving plan file
	runID string
	// flags for terraform apply
	isDestroy bool
}

type JobDBResult struct {
	JobID                  pgtype.Text `json:"job_id"`
	RunID                  pgtype.Text `json:"run_id"`
	JobType                pgtype.Name `json:"job_type"`
	Status                 pgtype.Text `json:"status"`
	IsDestroy              bool        `json:"is_destroy"`
	Refresh                bool        `json:"refresh"`
	RefreshOnly            bool        `json:"refresh_only"`
	AutoApply              bool        `json:"auto_apply"`
	Speculative            bool        `json:"speculative"`
	ConfigurationVersionID pgtype.Text `json:"configuration_version_id"`
	WorkspaceID            pgtype.Text `json:"workspace_id"`
}

type Action string

const (
	Confirmed Action = "confirmed"
)
