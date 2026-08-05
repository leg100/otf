package runner

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const KubeExecutorKind = "kubernetes"

func init() {
	homeDir, _ = os.UserHomeDir()

	defaultKubeConfig = kubeConfig{
		// Default to using the same version of the job image as the current
		// version of otfd.
		Image: fmt.Sprintf("leg100/otf-job:%s", internal.Version),
		// ConfigPath is only used as a fallback in case we aren't running
		// 'in-cluster', in which case it's assumed we're running on a
		// workstation for dev/testing purposes and there should be a home dir
		// and a kubectl config file.
		ConfigPath: filepath.Join(homeDir, ".kube", "config"),
		// OTF_KUBERNETES_NAMESPACE is set in the otfd helm chart to the value
		// of the current namespace otfd is deployed in.
		// If unset, this means otfd is probably running outside of a cluster,
		// in which case the namespace will be "", which is equivalent to the
		// 'default' namespace.
		Namespace: os.Getenv("OTF_KUBERNETES_NAMESPACE"),
		// OTF_KUBERNETES_SERVICE_ACCOUNT is set in the otfd helm chart to the value
		// of the current service account that otfd is running as.
		// If unset, this means otfd is probably running outside of a cluster,
		// in which case the job won't have an assigned service account.
		ServiceAccount: os.Getenv("OTF_KUBERNETES_SERVICE_ACCOUNT"),
		// OTF_KUBERNETES_CACHE_PVC is set in the otfd helm chart to the name of
		// the optional persistent volume claim for caching.
		CachePVC: os.Getenv("OTF_KUBERNETES_CACHE_PVC"),
		ServerURL: fmt.Sprintf("http://%s.%s:%s/",
			os.Getenv("OTF_KUBERNETES_SERVICE_NAME"),
			os.Getenv("OTF_KUBERNETES_NAMESPACE"),
			os.Getenv("OTF_KUBERNETES_SERVICE_PORT"),
		),
		// Delete job by default 1 hour after it has finished
		TTLAfterFinish: time.Hour,
		// Retry spawning a job this many times.
		SpawnRetries: 3,
		flags: kubeConfigFlags{
			RequestCPU:    "500m",
			RequestMemory: "128Mi",
		},
	}
}

var (
	homeDir           string
	defaultKubeConfig kubeConfig
)

type kubeConfig struct {
	Namespace      string
	Image          string
	ConfigPath     string
	ServerURL      string
	ServiceAccount string
	CachePVC       string
	TTLAfterFinish time.Duration
	SpawnRetries   int

	requestCPU    k8sresource.Quantity
	requestMemory k8sresource.Quantity
	limitCPU      *k8sresource.Quantity
	limitMemory   *k8sresource.Quantity
	labels        map[string]string

	flags kubeConfigFlags
}

// kubeConfigFlags are CLI flags that need to be parsed first and should not be
// used directly by the kubernetes executor.
type kubeConfigFlags struct {
	Labels        []string
	RequestCPU    string
	RequestMemory string
	LimitCPU      string
	LimitMemory   string
}

func registerKubeFlags(flags *pflag.FlagSet, cfg *kubeConfig) {
	flags.StringVar(&cfg.Image, "kubernetes-job-image", cfg.Image, "Image to use for kubernetes jobs.")
	flags.DurationVar(&cfg.TTLAfterFinish, "kubernetes-ttl-after-finish", cfg.TTLAfterFinish, "Delete finished kubernetes job after this duration.")
	flags.IntVar(&cfg.SpawnRetries, "kubernetes-spawn-retries", cfg.SpawnRetries, "Number of times to retry spawning a kubernetes job before reporting the job as errored. A retry resumes from the first step that has not yet completed. Zero disables retrying.")
	flags.StringVar(&cfg.flags.RequestCPU, "kubernetes-request-cpu", cfg.flags.RequestCPU, "Requested CPU for kubernetes job.")
	flags.StringVar(&cfg.flags.RequestMemory, "kubernetes-request-memory", cfg.flags.RequestMemory, "Requested memory for kubernetes job.")
	flags.StringVar(&cfg.flags.LimitCPU, "kubernetes-limit-cpu", cfg.flags.LimitCPU, "CPU limit for kubernetes job.")
	flags.StringVar(&cfg.flags.LimitMemory, "kubernetes-limit-memory", cfg.flags.LimitMemory, "Memory limit for kubernetes job.")
	flags.StringSliceVar(&cfg.flags.Labels, "kubernetes-labels", cfg.flags.Labels, "Set additional labels on kubernetes jobs. Name and value are separated by an equals sign, e.g. `foo=bar`.")
}

type kubeExecutor struct {
	Logger          logr.Logger
	Config          kubeConfig
	OperationConfig OperationConfig
	jobs            kubeExecutorJobsClient
	secrets         kubeExecutorSecretsClient
}

type kubeExecutorJobsClient interface {
	Create(ctx context.Context, job *batchv1.Job, opts metav1.CreateOptions) (*batchv1.Job, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*batchv1.Job, error)
	List(ctx context.Context, opts metav1.ListOptions) (*batchv1.JobList, error)
}

type kubeExecutorSecretsClient interface {
	Create(ctx context.Context, secret *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error)
	Update(ctx context.Context, secret *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error)
}

func newKubeExecutor(
	logger logr.Logger,
	operationConfig OperationConfig,
	kubeConfig kubeConfig,
) (*kubeExecutor, error) {
	executor := &kubeExecutor{
		Logger:          logger,
		OperationConfig: operationConfig,
		Config:          kubeConfig,
	}

	requestCPU, err := k8sresource.ParseQuantity(kubeConfig.flags.RequestCPU)
	if err != nil {
		return nil, fmt.Errorf("invalid cpu request quantity: %s: %w", kubeConfig.flags.RequestCPU, err)
	}
	executor.Config.requestCPU = requestCPU

	requestMemory, err := k8sresource.ParseQuantity(kubeConfig.flags.RequestMemory)
	if err != nil {
		return nil, fmt.Errorf("invalid memory request quantity: %s: %w", kubeConfig.flags.RequestMemory, err)
	}
	executor.Config.requestMemory = requestMemory

	if kubeConfig.flags.LimitCPU != "" {
		limitCPU, err := k8sresource.ParseQuantity(kubeConfig.flags.LimitCPU)
		if err != nil {
			return nil, fmt.Errorf("invalid cpu limit quantity: %s: %w", kubeConfig.flags.LimitCPU, err)
		}
		executor.Config.limitCPU = &limitCPU
	}

	if kubeConfig.flags.LimitMemory != "" {
		limitMemory, err := k8sresource.ParseQuantity(kubeConfig.flags.LimitMemory)
		if err != nil {
			return nil, fmt.Errorf("invalid memory limit quantity: %s: %w", kubeConfig.flags.LimitMemory, err)
		}
		executor.Config.limitMemory = &limitMemory
	}

	if kubeConfig.SpawnRetries < 0 {
		return nil, fmt.Errorf("invalid spawn retries: must not be negative: %d", kubeConfig.SpawnRetries)
	}

	executor.Config.labels = make(map[string]string)
	for _, label := range kubeConfig.flags.Labels {
		k, v, ok := strings.Cut(label, "=")
		if !ok {
			return nil, fmt.Errorf("invalid label: must be in format name=value")
		}
		executor.Config.labels[k] = v
	}

	// assume running in-cluster; otherwise use config path
	config, err := rest.InClusterConfig()
	if errors.Is(err, rest.ErrNotInCluster) {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig.ConfigPath)
	}
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes clientset: %w", err)
	}
	executor.secrets = clientset.CoreV1().Secrets(kubeConfig.Namespace)
	executor.jobs = clientset.BatchV1().Jobs(kubeConfig.Namespace)

	return executor, nil
}

// optionalStepError wraps an error from a spawn step that the kubernetes job
// does not depend upon. Such a step is retried like any other, but if the
// retries are exhausted the spawn as a whole must still be reported as having
// succeeded, because the job is running.
type optionalStepError struct{ err error }

func (e *optionalStepError) Error() string { return e.err.Error() }
func (e *optionalStepError) Unwrap() error { return e.err }

// spawn is an attempt-in-progress at spawning a kubernetes job for an OTF job.
// Its specs are built once and then held fixed - in particular the generated
// name, so that a retry addresses the same objects rather than creating a second
// secret and a second kubernetes job, which would carry out the OTF job twice
// concurrently. The remaining fields record progress, so that a retry resumes
// from the first step that has not yet completed.
type spawn struct {
	job    *Job
	name   string
	secret *corev1.Secret
	spec   *batchv1.Job

	secretCreated bool
	jobCreated    bool
	kjob          *batchv1.Job // nil until the created job's UID is known
	ownerRefSet   bool
}

// retry invokes fn, retrying with exponential backoff up to SpawnRetries times
// before returning the last error.
func (s *kubeExecutor) retry(ctx context.Context, sp *spawn, fn func() error) error {
	if s.Config.SpawnRetries <= 0 {
		// Retrying is disabled: make a single attempt. NOTE: this must be
		// handled explicitly because backoff.WithMaxRetries treats a maximum of
		// zero as unlimited.
		return fn()
	}
	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = 500 * time.Millisecond
	policy.MaxInterval = 5 * time.Second
	// Bound the retries solely by their number, not by elapsed time: an attempt
	// can itself block for the duration of the API server's etcd request
	// timeout, which would make the effective number of attempts unpredictable.
	policy.MaxElapsedTime = 0

	tries := backoff.WithMaxRetries(policy, uint64(s.Config.SpawnRetries))
	return backoff.RetryNotify(
		fn,
		backoff.WithContext(tries, ctx),
		func(err error, next time.Duration) {
			s.Logger.Error(err, "retrying spawn of kubernetes job", "name", sp.name, "otf-job", sp.job, "backoff", next)
		},
	)
}

// spawn makes the kubernetes API calls needed to spawn a job, resuming from the
// first step that has not yet completed.
// Steps that the kubernetes job does not depend upon return an optionalStepError.
func (s *kubeExecutor) spawn(ctx context.Context, sp *spawn) error {
	if !sp.secretCreated {
		_, err := s.secrets.Create(ctx, sp.secret, metav1.CreateOptions{})
		switch {
		case apierrors.IsAlreadyExists(err):
			// An earlier attempt timed out but did in fact create the secret.
		case err != nil:
			return fmt.Errorf("creating kubernetes secret for job token: %w", err)
		}
		sp.secretCreated = true
		s.Logger.V(4).Info("created kubernetes secret for job token", "name", sp.name, "namespace", s.Config.Namespace, "otf-job", sp.job)
	}

	if !sp.jobCreated {
		kjob, err := s.jobs.Create(ctx, sp.spec, metav1.CreateOptions{})
		switch {
		case apierrors.IsAlreadyExists(err):
			// Earlier attempt created the job; UID is unknown and to be retrieved below.
		case err != nil:
			return fmt.Errorf("creating kubernetes job: %w", err)
		default:
			sp.kjob = kjob
		}
		sp.jobCreated = true
		s.Logger.V(1).Info("created kubernetes job", "name", sp.name, "namespace", s.Config.Namespace, "otf-job", sp.job)
	}

	// The kubernetes job now exists and is going to run to completion, so every
	// remaining step is optional.

	if sp.kjob == nil {
		// The create response was lost to a timeout. Retrieve the job in order
		// to obtain the UID needed for the owner reference.
		kjob, err := s.jobs.Get(ctx, sp.name, metav1.GetOptions{})
		if err != nil {
			return &optionalStepError{fmt.Errorf("retrieving kubernetes job created by an earlier attempt: %w", err)}
		}
		sp.kjob = kjob
	}

	if !sp.ownerRefSet {
		// Set secret's owner to its job, so that it is deleted when its job is deleted.
		sp.secret.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: "batch/v1",
				Kind:       "job",
				Name:       sp.kjob.Name,
				UID:        sp.kjob.UID,
			},
		}
		if _, err := s.secrets.Update(ctx, sp.secret, metav1.UpdateOptions{}); err != nil {
			return &optionalStepError{fmt.Errorf("setting kubernetes job token secret owner reference: %w", err)}
		}
		sp.ownerRefSet = true
	}

	return nil
}

func (s *kubeExecutor) SpawnOperation(ctx context.Context, _ *errgroup.Group, job *Job, jobToken []byte) error {
	labels := map[string]string{
		"app.kubernetes.io/name":     "otf-job",
		"app.kubernetes.io/instance": job.ID.String(),
		"app.kubernetes.io/version":  internal.Version,
		"app.kubernetes.io/part-of":  "otf",
		"otf.ninja/job-id":           job.ID.String(),
		"otf.ninja/run-id":           job.RunID.String(),
		"otf.ninja/runner-id":        job.RunnerID.String(),
		"otf.ninja/workspace-id":     job.WorkspaceID.String(),
		"otf.ninja/organization":     job.Organization.String(),
	}
	maps.Copy(labels, s.Config.labels)

	const (
		cacheVolumeName   = "cache"
		jobTokenSecretKey = "jobToken"
	)

	// Generate name for k8s job. OTF uses TFE IDs, which use the base58 alphabet, which
	// includes upper case letters; but they're not permissible in k8s resource
	// names. Instead we lower case the OTF job TFE ID and add a random suffix
	// to reduce the possibility of duplicate names (the risk of which by using lower cased
	// base58 is slightly increased).
	//
	// We could have instead used k8s' GenerateName func, which generates a
	// random name on the server side, but we would like more control over the
	// format of the generated name.
	lowerCaseAndNumbers := "abcdefghijkmnopqrstuvwxyz0123456789"
	suffix := internal.GenerateRandomStringFromAlphabet(4, lowerCaseAndNumbers)
	jobName := fmt.Sprintf("%s-%s", strings.ToLower(job.ID.String()), suffix)

	// Create secret containing job token.
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: s.Config.Namespace,
			Labels:    labels,
		},
		Immutable: new(true),
		StringData: map[string]string{
			jobTokenSecretKey: string(jobToken),
		},
	}
	// Create k8s job for OTF job.
	spec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: s.Config.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			// A job by default will re-create pods upon failure (up to 6 times
			// with backoff), but we can't guarantee idempotency.
			BackoffLimit:            new(int32(0)),
			TTLSecondsAfterFinished: new(int32(s.Config.TTLAfterFinish.Seconds())),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: s.Config.ServiceAccount,
					RestartPolicy:      corev1.RestartPolicyNever,
					Resources: &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    s.Config.requestCPU,
							corev1.ResourceMemory: s.Config.requestMemory,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "otf-job",
							Image: s.Config.Image,
							Env: []corev1.EnvVar{
								{
									Name:  "OTF_URL",
									Value: s.Config.ServerURL,
								},
								{
									Name:  "OTF_JOB_ID",
									Value: job.ID.String(),
								},
								{
									Name: "OTF_JOB_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: jobName,
											},
											Key: jobTokenSecretKey,
										},
									},
								},
								{
									Name:  "OTF_V",
									Value: strconv.Itoa(s.Logger.Verbosity),
								},
								{
									Name:  "OTF_LOG_FORMAT",
									Value: string(s.Logger.Format),
								},
								{
									Name:  "OTF_ENGINE_BINS_DIR",
									Value: s.OperationConfig.EngineBinDir,
								},
								{
									Name:  "OTF_PLUGIN_CACHE",
									Value: strconv.FormatBool(s.OperationConfig.PluginCache),
								},
								{
									Name:  "OTF_PLUGIN_CACHE_DIR",
									Value: s.OperationConfig.PluginCacheDir,
								},
								{
									Name:  "OTF_DEBUG",
									Value: strconv.FormatBool(s.OperationConfig.Debug),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      cacheVolumeName,
									MountPath: s.OperationConfig.EngineBinDir,
									SubPath:   filepath.Base(s.OperationConfig.EngineBinDir),
								},
								{
									Name:      cacheVolumeName,
									MountPath: s.OperationConfig.PluginCacheDir,
									SubPath:   filepath.Base(s.OperationConfig.PluginCacheDir),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: cacheVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	// Populate resource limits if user has specified any.
	if s.Config.limitCPU != nil || s.Config.limitMemory != nil {
		spec.Spec.Template.Spec.Resources.Limits = corev1.ResourceList{}
	}
	if s.Config.limitCPU != nil {
		spec.Spec.Template.Spec.Resources.Limits[corev1.ResourceCPU] = *s.Config.limitCPU
	}
	if s.Config.limitMemory != nil {
		spec.Spec.Template.Spec.Resources.Limits[corev1.ResourceMemory] = *s.Config.limitMemory
	}

	if s.Config.CachePVC != "" {
		spec.Spec.Template.Spec.Volumes[0].VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: s.Config.CachePVC,
			},
		}
	}
	// Retry the spawn as a whole rather than each API call individually: an
	// attempt resumes from the first step that has not yet completed, so a retry
	// costs no more API calls than retrying that step alone would, but there is
	// a single budget and a single place where the outcome is decided.
	sp := &spawn{job: job, name: jobName, secret: secret, spec: spec}
	err := s.retry(ctx, sp, func() error { return s.spawn(ctx, sp) })

	var optional *optionalStepError
	if errors.As(err, &optional) {
		s.Logger.Error(optional, "spawned kubernetes job but could not complete an optional step; job token secret will not be garbage collected with its job", "name", jobName, "namespace", s.Config.Namespace, "otf-job", job)
		return nil
	}
	return err
}

func (s *kubeExecutor) currentJobs(ctx context.Context, runnerID resource.TfeID) int {
	jobs, err := s.jobs.List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("otf.ninja/runner-id=%s", runnerID),
	})
	if err != nil {
		s.Logger.Error(err, "listing current number of kubernetes jobs")
	}
	return len(jobs.Items)
}
