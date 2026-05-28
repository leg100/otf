package runner

import (
	"context"
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

func TestKubeExecutor_SpawnOperation_Hostname(t *testing.T) {
	containerEnv := func(t *testing.T, cfg kubeConfig) []corev1.EnvVar {
		t.Helper()
		executor, err := newKubeExecutor(logr.Discard(), defaultOperationConfig(), cfg)
		require.NoError(t, err)

		jobsClient := &fakeJobsClient{}
		executor.jobs = jobsClient
		executor.secrets = &fakeSecretsClient{}

		job := &Job{
			ID:           resource.NewTfeID(resource.JobKind),
			RunID:        resource.NewTfeID(resource.RunKind),
			Phase:        run.PlanPhase,
			Status:       JobAllocated,
			Organization: organization.NewTestName(t),
			WorkspaceID:  resource.NewTfeID(resource.WorkspaceKind),
			RunnerID:     new(resource.NewTfeID(resource.RunnerKind)),
		}
		require.NoError(t, executor.SpawnOperation(t.Context(), nil, job, []byte("token")))
		return jobsClient.job.Spec.Template.Spec.Containers[0].Env
	}

	envValue := func(envs []corev1.EnvVar, name string) string {
		for _, e := range envs {
			if e.Name == name {
				return e.Value
			}
		}
		return ""
	}

	t.Run("propagates public hostname to OTF_HOSTNAME", func(t *testing.T) {
		cfg := defaultKubeConfig
		cfg.Hostname = "app.otf.example.com"
		envs := containerEnv(t, cfg)
		assert.Equal(t, "app.otf.example.com", envValue(envs, "OTF_HOSTNAME"))
	})

	t.Run("OTF_HOSTNAME is empty when Hostname not configured", func(t *testing.T) {
		cfg := defaultKubeConfig
		cfg.Hostname = ""
		envs := containerEnv(t, cfg)
		assert.Equal(t, "", envValue(envs, "OTF_HOSTNAME"))
	})
}

type fakeSecretsClient struct {
	secret *corev1.Secret
}

func (f *fakeSecretsClient) Create(ctx context.Context, secret *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error) {
	f.secret = secret
	return secret, nil
}

func (f *fakeSecretsClient) Update(ctx context.Context, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
	f.secret = secret
	return secret, nil
}

type fakeJobsClient struct {
	job *batchv1.Job
}

func (f *fakeJobsClient) Create(ctx context.Context, job *batchv1.Job, opts metav1.CreateOptions) (*batchv1.Job, error) {
	f.job = job
	return job, nil
}

func (f *fakeJobsClient) List(ctx context.Context, opts metav1.ListOptions) (*batchv1.JobList, error) {
	return &batchv1.JobList{Items: []batchv1.Job{*f.job}}, nil
}
