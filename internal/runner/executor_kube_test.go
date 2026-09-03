package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewKubeExecutor(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		_, err := newKubeExecutor(
			logr.Discard(),
			defaultOperationConfig(),
			defaultKubeConfig,
		)
		require.NoError(t, err)
	})

	t.Run("with resource limits", func(t *testing.T) {
		cfg := defaultKubeConfig
		cfg.flags.LimitCPU = "3000m"
		cfg.flags.LimitMemory = "512Mi"

		_, err := newKubeExecutor(
			logr.Discard(),
			defaultOperationConfig(),
			cfg,
		)
		require.NoError(t, err)
	})

	t.Run("with invalid resource limits", func(t *testing.T) {
		cfg := defaultKubeConfig
		cfg.flags.LimitCPU = "foo"
		cfg.flags.LimitMemory = "bar"

		_, err := newKubeExecutor(
			logr.Discard(),
			defaultOperationConfig(),
			cfg,
		)
		assert.Error(t, err)
	})

	t.Run("with labels", func(t *testing.T) {
		cfg := defaultKubeConfig
		cfg.flags.Labels = []string{"foo=bar", "coo=boo"}

		_, err := newKubeExecutor(
			logr.Discard(),
			defaultOperationConfig(),
			cfg,
		)
		require.NoError(t, err)
	})

	t.Run("with invalid labels", func(t *testing.T) {
		cfg := defaultKubeConfig
		cfg.flags.Labels = []string{"foobar", "cooboo"}

		_, err := newKubeExecutor(
			logr.Discard(),
			defaultOperationConfig(),
			cfg,
		)
		assert.Error(t, err)
	})
}

func TestKubeExecutor_SpawnOperation(t *testing.T) {
	cfg := defaultKubeConfig
	cfg.flags.Labels = []string{"foo=bar"}
	cfg.flags.LimitCPU = "3000m"
	cfg.flags.LimitMemory = "512Mi"

	executor, err := newKubeExecutor(
		logr.Discard(),
		defaultOperationConfig(),
		cfg,
	)
	require.NoError(t, err)

	jobsClient := &fakeJobsClient{}
	executor.jobs = jobsClient

	secretsClient := &fakeSecretsClient{}
	executor.secrets = secretsClient

	job := &Job{
		ID:           resource.NewTfeID(resource.JobKind),
		RunID:        resource.NewTfeID(resource.RunKind),
		Phase:        run.PlanPhase,
		Status:       JobAllocated,
		Organization: organization.NewTestName(t),
		WorkspaceID:  resource.NewTfeID(resource.WorkspaceKind),
		RunnerID:     new(resource.NewTfeID(resource.RunnerKind)),
	}

	err = executor.SpawnOperation(t.Context(), nil, job, []byte("token"))
	require.NoError(t, err)

	wantLabels := map[string]string{
		"app.kubernetes.io/instance": job.ID.String(),
		"app.kubernetes.io/name":     "otf-job",
		"app.kubernetes.io/part-of":  "otf",
		"app.kubernetes.io/version":  "unknown",
		"otf.ninja/job-id":           job.ID.String(),
		"otf.ninja/organization":     job.Organization.String(),
		"otf.ninja/run-id":           job.RunID.String(),
		"otf.ninja/runner-id":        job.RunnerID.String(),
		"otf.ninja/workspace-id":     job.WorkspaceID.String(),
		"foo":                        "bar",
	}
	assert.Equal(t, wantLabels, jobsClient.job.Labels)
	assert.Equal(t, wantLabels, secretsClient.secret.Labels)
	assert.Equal(t, map[string]string{"jobToken": "token"}, secretsClient.secret.StringData)
}

type fakeSecretsClient struct {
	secret    *corev1.Secret
	deleted   []string
	updateErr error
}

func (f *fakeSecretsClient) Create(ctx context.Context, secret *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error) {
	f.secret = secret
	return secret, nil
}

func (f *fakeSecretsClient) Update(ctx context.Context, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	f.secret = secret
	return secret, nil
}

func (f *fakeSecretsClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	f.deleted = append(f.deleted, name)
	return nil
}

type fakeJobsClient struct {
	job       *batchv1.Job
	createErr error
	listErr   error
}

func (f *fakeJobsClient) Create(ctx context.Context, job *batchv1.Job, opts metav1.CreateOptions) (*batchv1.Job, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.job = job
	return job, nil
}

func (f *fakeJobsClient) List(ctx context.Context, opts metav1.ListOptions) (*batchv1.JobList, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return &batchv1.JobList{Items: []batchv1.Job{*f.job}}, nil
}

// TestKubeExecutor_SpawnOperationDeletesOrphanedSecret checks that the secret
// containing the job token is deleted when it has been left without an owner to
// garbage collect it.
func TestKubeExecutor_SpawnOperationDeletesOrphanedSecret(t *testing.T) {
	newTestJob := func() *Job {
		return &Job{
			ID:           resource.NewTfeID(resource.JobKind),
			RunID:        resource.NewTfeID(resource.RunKind),
			Phase:        run.PlanPhase,
			Status:       JobAllocated,
			Organization: organization.NewTestName(t),
			WorkspaceID:  resource.NewTfeID(resource.WorkspaceKind),
			RunnerID:     new(resource.NewTfeID(resource.RunnerKind)),
		}
	}

	t.Run("job creation fails", func(t *testing.T) {
		secretsClient := &fakeSecretsClient{}
		executor := &kubeExecutor{
			Logger:  logr.Discard(),
			jobs:    &fakeJobsClient{createErr: errors.New("job quota exceeded")},
			secrets: secretsClient,
		}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(), []byte("token"))
		require.Error(t, err)
		assert.Len(t, secretsClient.deleted, 1)
	})

	t.Run("setting owner reference fails", func(t *testing.T) {
		secretsClient := &fakeSecretsClient{updateErr: errors.New("conflict")}
		executor := &kubeExecutor{
			Logger:  logr.Discard(),
			jobs:    &fakeJobsClient{},
			secrets: secretsClient,
		}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(), []byte("token"))
		require.Error(t, err)
		assert.Len(t, secretsClient.deleted, 1)
	})
}

// TestKubeExecutor_CurrentJobsListError checks that a failure to list jobs is
// reported as zero jobs rather than dereferencing the nil job list.
func TestKubeExecutor_CurrentJobsListError(t *testing.T) {
	executor := &kubeExecutor{
		Logger: logr.Discard(),
		jobs:   &fakeJobsClient{listErr: errors.New("api server unavailable")},
	}

	assert.Equal(t, 0, executor.currentJobs(t.Context(), resource.NewTfeID(resource.RunnerKind)))
}
