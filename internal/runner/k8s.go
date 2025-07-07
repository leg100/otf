package runner

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/logr"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeOperationSpawner struct {
	Config    Config
	Logger    logr.Logger
	URL       string
	Namespace string

	kube *kubernetes.Clientset
}

func NewKubeOperationSpawner(logger logr.Logger, cfg Config, url string) (*KubeOperationSpawner, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes config: %w", err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes clientset: %w", err)
	}
	return &KubeOperationSpawner{
		Config: cfg,
		Logger: logger,
		URL:    url,
		kube:   clientset,
	}, nil
}

func (s *KubeOperationSpawner) SpawnOperation(ctx context.Context, _ *errgroup.Group, job *Job, jobToken []byte) error {
	// Launch k8s job
	spec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.String(),
			Namespace: s.Namespace,
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
							Image: "",
						},
					},
				},
			},
		},
	}
	_, err := s.kube.BatchV1().Jobs(s.Namespace).Create(ctx, spec, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
