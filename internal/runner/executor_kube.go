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

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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
	flags.DurationVar(&cfg.TTLAfterFinish, "kubernetes-ttl-after-finish", cfg.TTLAfterFinish, "Delete finished kubernetes job after this duration.")
	flags.StringVar(&cfg.flags.RequestCPU, "kubernetes-request-cpu", cfg.flags.RequestCPU, "Requested CPU for kubernetes job.")
	flags.StringVar(&cfg.flags.RequestMemory, "kubernetes-request-memory", cfg.flags.RequestMemory, "Requested memory for kubernetes job.")
	flags.StringVar(&cfg.flags.LimitCPU, "kubernetes-limit-cpu", cfg.flags.LimitCPU, "CPU limit for kubernetes job.")
	flags.StringVar(&cfg.flags.LimitMemory, "kubernetes-limit-memory", cfg.flags.LimitMemory, "Memory limit for kubernetes job.")
	flags.StringSliceVar(&cfg.flags.Labels, "kubernetes-labels", cfg.flags.Labels, "Labels to add to kubernetes job.")
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
	ksecret, err := s.secrets.Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating kubernetes secret for job token: %w", err)
	}
	s.Logger.V(4).Info("created kubernetes secret for job token", "name", ksecret.GetName(), "namespace", ksecret.GetNamespace(), "otf-job", job)

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
												Name: ksecret.GetName(),
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
	kjob, err := s.jobs.Create(ctx, spec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating kubernetes job: %w", err)
	}
	s.Logger.V(1).Info("created kubernetes job", "name", kjob.GetName(), "namespace", kjob.GetNamespace(), "otf-job", job)

	// Set secret's owner to its job, so that it is deleted when its job is
	// deleted.
	secret.OwnerReferences = []metav1.OwnerReference{
		{
			// NOTE: the API version and kind are empty strings in the returned
			// job struct, so we're forced to hardcode them.
			APIVersion: "batch/v1",
			Kind:       "job",
			Name:       kjob.Name,
			UID:        kjob.UID,
		},
	}
	_, err = s.secrets.Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("setting kubernetes job token secret owner reference: %w", err)
	}

	return nil
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
