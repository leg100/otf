package runner

import (
	"context"
	"fmt"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
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

func (s *KubeOperationSpawner) NewOperation(ctx context.Context, job *Job, jobToken []byte) (*operation, error) {
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
		return nil, err
	}
	client, err := otfapi.NewClient(otfapi.Config{
		URL:           s.URL,
		Token:         string(jobToken),
		RetryRequests: true,
		Logger:        s.Logger,
	})
	if err != nil {
		return nil, err
	}
	return newOperation(ctx, operationOptions{
		logger:          s.Logger,
		OperationConfig: s.Config.OperationConfig,
		job:             job,
		jobToken:        jobToken,
		runs:            &run.Client{Client: client},
		jobs:            &remoteClient{Client: client},
		workspaces:      &workspace.Client{Client: client},
		variables:       &variable.Client{Client: client},
		state:           &state.Client{Client: client},
		configs:         &configversion.Client{Client: client},
		server:          client,
	})
}
