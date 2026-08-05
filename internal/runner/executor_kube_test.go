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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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

// TestKubeExecutor_SpawnOperation_retry tests the handling of the transient
// kubernetes API errors that occur when its etcd backend is under load. A job
// that fails to spawn cannot be spawned later, so a transient error must be
// retried, and an error arising *after* the kubernetes job has been created must
// not be reported as a failure to spawn.
func TestKubeExecutor_SpawnOperation_retry(t *testing.T) {
	// etcdTimeout is representative of the error the kubernetes API server
	// returns when its etcd backend cannot commit a write in time.
	etcdTimeout := errors.New("etcdserver: request timed out")

	newExecutor := func(t *testing.T, retries int) (*kubeExecutor, *fakeSecretsClient, *fakeJobsClient) {
		t.Helper()
		cfg := defaultKubeConfig
		cfg.SpawnRetries = retries
		executor, err := newKubeExecutor(logr.Discard(), defaultOperationConfig(), cfg)
		require.NoError(t, err)
		secrets := &fakeSecretsClient{}
		jobs := &fakeJobsClient{}
		executor.secrets = secrets
		executor.jobs = jobs
		return executor, secrets, jobs
	}
	newTestJob := func(t *testing.T) *Job {
		t.Helper()
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

	t.Run("retries transient secret create error", func(t *testing.T) {
		executor, secrets, jobs := newExecutor(t, 3)
		secrets.createErrs = []error{etcdTimeout, etcdTimeout}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.NoError(t, err)
		assert.Equal(t, 3, secrets.creates, "should have created secret on third attempt")
		assert.Equal(t, 1, jobs.creates, "should have created job once")
	})

	t.Run("retries transient job create error", func(t *testing.T) {
		executor, _, jobs := newExecutor(t, 3)
		jobs.createErrs = []error{etcdTimeout}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.NoError(t, err)
		assert.Equal(t, 2, jobs.creates, "should have created job on second attempt")
	})

	t.Run("gives up after exhausting retries", func(t *testing.T) {
		executor, secrets, jobs := newExecutor(t, 1)
		secrets.createErrs = []error{etcdTimeout, etcdTimeout}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.Error(t, err)
		assert.ErrorIs(t, err, etcdTimeout)
		assert.Equal(t, 2, secrets.creates, "should have made one attempt plus one retry")
		assert.Zero(t, jobs.creates, "should not have created job")
	})

	t.Run("makes single attempt when retrying is disabled", func(t *testing.T) {
		executor, secrets, _ := newExecutor(t, 0)
		secrets.createErrs = []error{etcdTimeout}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.Error(t, err)
		assert.Equal(t, 1, secrets.creates)
	})

	t.Run("tolerates secret already existing", func(t *testing.T) {
		executor, secrets, jobs := newExecutor(t, 3)
		// An earlier attempt timed out but did in fact create the secret.
		secrets.createErrs = []error{apierrors.NewAlreadyExists(schema.GroupResource{Resource: "secrets"}, "job-abc")}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.NoError(t, err)
		assert.Equal(t, 1, secrets.creates, "should not have retried")
		assert.Equal(t, 1, jobs.creates, "should have proceeded to create job")
	})

	t.Run("tolerates job already existing", func(t *testing.T) {
		executor, secrets, jobs := newExecutor(t, 3)
		jobs.createErrs = []error{apierrors.NewAlreadyExists(schema.GroupResource{Resource: "jobs"}, "job-abc")}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.NoError(t, err)
		assert.Equal(t, 1, jobs.creates, "should not have retried")
		// The create response was lost, so the job is fetched to recover the UID
		// needed for the owner reference.
		assert.Equal(t, 1, jobs.gets, "should have fetched the existing job")
		assert.Equal(t, 1, secrets.creates, "should not have re-created the secret")
		require.Len(t, secrets.secret.OwnerReferences, 1)
		assert.Equal(t, types.UID("fake-uid"), secrets.secret.OwnerReferences[0].UID)
	})

	t.Run("owner reference is skipped if existing job cannot be fetched", func(t *testing.T) {
		executor, secrets, jobs := newExecutor(t, 3)
		jobs.createErrs = []error{apierrors.NewAlreadyExists(schema.GroupResource{Resource: "jobs"}, "job-abc")}
		jobs.getErrs = []error{etcdTimeout, etcdTimeout, etcdTimeout, etcdTimeout}

		// The job is running regardless, so this must not fail the spawn.
		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.NoError(t, err)
		assert.Equal(t, 4, jobs.gets, "should have retried on the shared spawn budget")
		assert.Equal(t, 1, secrets.creates, "completed steps should not be repeated")
		assert.Equal(t, 1, jobs.creates, "completed steps should not be repeated")
		assert.Zero(t, secrets.updates, "should not have attempted to set owner reference")
	})

	t.Run("retries transient owner reference error", func(t *testing.T) {
		executor, secrets, jobs := newExecutor(t, 3)
		secrets.updateErrs = []error{etcdTimeout}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.NoError(t, err)
		assert.Equal(t, 2, secrets.updates, "should have set owner reference on second attempt")
		assert.NotEmpty(t, secrets.secret.OwnerReferences, "secret should not be left orphaned")
		assert.Equal(t, 1, secrets.creates, "should not have re-created the secret")
		assert.Equal(t, 1, jobs.creates, "should not have re-created the job")
	})

	t.Run("owner reference failure does not fail the spawn", func(t *testing.T) {
		executor, secrets, jobs := newExecutor(t, 3)
		secrets.updateErrs = []error{etcdTimeout, etcdTimeout, etcdTimeout, etcdTimeout}

		// The kubernetes job has been created and is going to run, so the job
		// must not be reported as having failed to spawn - the secret is leaked
		// instead, which is by far the lesser evil.
		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.NoError(t, err)
		assert.Equal(t, 4, secrets.updates, "should have retried on the shared spawn budget")
		assert.Equal(t, 1, secrets.creates, "completed steps should not be repeated")
		assert.Equal(t, 1, jobs.creates, "completed steps should not be repeated")
	})

	t.Run("owner reference is not retried when retrying is disabled", func(t *testing.T) {
		executor, secrets, _ := newExecutor(t, 0)
		secrets.updateErrs = []error{etcdTimeout}

		err := executor.SpawnOperation(t.Context(), nil, newTestJob(t), []byte("token"))
		require.NoError(t, err)
		assert.Equal(t, 1, secrets.updates)
	})
}

type fakeSecretsClient struct {
	secret *corev1.Secret
	// createErrs are returned by successive calls to Create; a nil entry, or
	// exhausting the slice, means the call succeeds.
	createErrs []error
	creates    int
	// updateErrs are returned by successive calls to Update; a nil entry, or
	// exhausting the slice, means the call succeeds.
	updateErrs []error
	updates    int
}

func (f *fakeSecretsClient) Create(ctx context.Context, secret *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error) {
	f.creates++
	if len(f.createErrs) > 0 {
		err := f.createErrs[0]
		f.createErrs = f.createErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	f.secret = secret
	return secret, nil
}

func (f *fakeSecretsClient) Update(ctx context.Context, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
	f.updates++
	if len(f.updateErrs) > 0 {
		err := f.updateErrs[0]
		f.updateErrs = f.updateErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	f.secret = secret
	return secret, nil
}

type fakeJobsClient struct {
	job *batchv1.Job
	// createErrs are returned by successive calls to Create; a nil entry, or
	// exhausting the slice, means the call succeeds.
	createErrs []error
	creates    int
	// getErrs are returned by successive calls to Get; a nil entry, or
	// exhausting the slice, means the call succeeds.
	getErrs []error
	gets    int
}

func (f *fakeJobsClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*batchv1.Job, error) {
	f.gets++
	if len(f.getErrs) > 0 {
		err := f.getErrs[0]
		f.getErrs = f.getErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name, UID: "fake-uid"},
	}, nil
}

func (f *fakeJobsClient) Create(ctx context.Context, job *batchv1.Job, opts metav1.CreateOptions) (*batchv1.Job, error) {
	f.creates++
	if len(f.createErrs) > 0 {
		err := f.createErrs[0]
		f.createErrs = f.createErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	f.job = job
	return job, nil
}

func (f *fakeJobsClient) List(ctx context.Context, opts metav1.ListOptions) (*batchv1.JobList, error) {
	if f.job == nil {
		return &batchv1.JobList{}, nil
	}
	return &batchv1.JobList{Items: []batchv1.Job{*f.job}}, nil
}
