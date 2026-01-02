package runner

import (
	"context"
	"fmt"
	"strconv"

	"github.com/leg100/otf/internal/logr"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const kubeExecutorKind = "kubernetes"

type KubeConfig struct {
	Namespace string
}

type kubeExecutor struct {
	Logger          logr.Logger
	URL             string
	Config          KubeConfig
	OperationConfig OperationConfig

	kube *kubernetes.Clientset
}

func newKubeExecutor(logger logr.Logger, cfg OperationConfig, url string) (*kubeExecutor, error) {
	config, err := rest.InClusterConfig()
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
		OperationConfig: cfg,
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
			Name:      job.String(),
			Namespace: s.Config.Namespace,
			Labels: map[string]string{
				"job-id":       job.ID.String(),
				"runner-id":    job.RunnerID.String(),
				"workspace-id": job.WorkspaceID.String(),
				"organization": job.Organization.String(),
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "otf-job",
							Image: "leg100/otfd:latest",
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
									Value: strconv.Itoa(s.Logger.GetV()),
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
		return err
	}
	return nil
}

func (s *kubeExecutor) currentJobs() int {
	// TODO: list k8s jobs match runner
	return 0
}
