package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const KubeExecutorKind = "kubernetes"

func init() {
	homeDir, _ = os.UserHomeDir()

	defaultKubeConfig = &kubeConfig{
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
		ServerURL: &KubeServerURLFlag{
			url: fmt.Sprintf("http://%s.%s:%s/",
				os.Getenv("OTF_KUBERNETES_SERVICE_NAME"),
				os.Getenv("OTF_KUBERNETES_NAMESPACE"),
				os.Getenv("OTF_KUBERNETES_SERVICE_PORT"),
			),
		},
	}
}

var (
	homeDir           string
	defaultKubeConfig *kubeConfig
)

type kubeConfig struct {
	Namespace      string
	Image          string
	ConfigPath     string
	ServerURL      KubeConfigServerURL
	ServiceAccount string
	CachePVC       string
}

type KubeConfigServerURL interface {
	String() string
}

type kubeExecutor struct {
	Logger          logr.Logger
	Config          kubeConfig
	OperationConfig OperationConfig

	kube *kubernetes.Clientset
}

func newKubeExecutor(
	logger logr.Logger,
	operationConfig OperationConfig,
	kubeConfig kubeConfig,
) (*kubeExecutor, error) {
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
	return &kubeExecutor{
		Logger:          logger,
		OperationConfig: operationConfig,
		Config:          kubeConfig,
		kube:            clientset,
	}, nil
}

func (s *kubeExecutor) SpawnOperation(ctx context.Context, _ *errgroup.Group, job *Job, jobToken []byte) error {
	// Launch k8s job
	//
	// TODO:
	// * support optional persistent volumes for:
	// 	* engine binaries
	// 	* provider cache (will opentofu concurrency support work?)
	labels := map[string]string{
		"otf.ninja/job-id":       job.ID.String(),
		"otf.ninja/run-id":       job.RunID.String(),
		"otf.ninja/runner-id":    job.RunnerID.String(),
		"otf.ninja/workspace-id": job.WorkspaceID.String(),
		"otf.ninja/organization": job.Organization.String(),
	}
	spec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: strings.ToLower(job.ID.String()),
			Namespace:    s.Config.Namespace,
			Labels:       labels,
		},
		Spec: batchv1.JobSpec{
			// A job by default will re-create pods upon failure (up to 6 times
			// with backoff), but we can't guarantee idempotency.
			BackoffLimit: internal.Ptr(int32(0)),
			// Delete job an hour after it has finished.
			TTLSecondsAfterFinished: internal.Ptr(int32(time.Hour.Seconds())),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: s.Config.ServiceAccount,
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "otf-job",
							Image: s.Config.Image,
							Env: []corev1.EnvVar{
								{
									Name:  "OTF_URL",
									Value: s.Config.ServerURL.String(),
								},
								{
									Name:  "OTF_JOB_ID",
									Value: job.ID.String(),
								},
								{
									Name:  "OTF_JOB_TOKEN",
									Value: string(jobToken),
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
									Name:      "cache",
									MountPath: "/cache",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "cache",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
	if s.Config.CachePVC != "" {
		spec.Spec.Template.Spec.Volumes[0].VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: s.Config.CachePVC,
			},
		}
	}

	kjob, err := s.kube.BatchV1().Jobs(s.Config.Namespace).Create(ctx, spec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating kubernetes job: %w", err)
	}
	s.Logger.V(1).Info("created kubernetes job", "name", kjob.GetGenerateName(), "namespace", kjob.GetNamespace(), "otf-job", job)
	return nil
}

func (s *kubeExecutor) currentJobs() int {
	// TODO: list k8s jobs match runner
	return 0
}
