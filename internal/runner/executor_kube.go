package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	defaultKubeConfig = kubeConfig{
		Image:      "leg100/otf-job:" + internal.Version,
		ConfigPath: filepath.Join(homeDir, ".kube", "config"),
		Namespace:  "default",
	}
}

var (
	homeDir           string
	defaultKubeConfig kubeConfig
)

type kubeConfig struct {
	Namespace  string
	Image      string
	ConfigPath string
}

type kubeExecutor struct {
	Logger          logr.Logger
	URL             string
	Config          kubeConfig
	OperationConfig OperationConfig

	kube *kubernetes.Clientset
}

func newKubeExecutor(
	logger logr.Logger,
	operationConfig OperationConfig,
	kubeConfig kubeConfig,
	url string,
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
		URL:             url,
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
	spec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: strings.ToLower(job.ID.String()),
			Namespace:    s.Config.Namespace,
			Labels: map[string]string{
				"job-id":       job.ID.String(),
				"runner-id":    job.RunnerID.String(),
				"workspace-id": job.WorkspaceID.String(),
				"organization": job.Organization.String(),
			},
		},
		Spec: batchv1.JobSpec{
			// A job by default will re-create pods upon failure (up to 6 times
			// with backoff), but we can't guarantee idempotency.
			BackoffLimit: internal.Ptr(int32(0)),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "otf-job",
							Image: s.Config.Image,
							Env: []corev1.EnvVar{
								{
									Name:  "OTF_URL",
									Value: s.URL,
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
							},
						},
					},
				},
			},
		},
	}
	_, err := s.kube.BatchV1().Jobs(s.Config.Namespace).Create(ctx, spec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating kubernetes job: %w", err)
	}
	return nil
}

func (s *kubeExecutor) currentJobs() int {
	// TODO: list k8s jobs match runner
	return 0
}
