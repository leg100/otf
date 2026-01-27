package integration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8swait "k8s.io/apimachinery/pkg/util/wait"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/goccy/go-yaml"

	"github.com/hashicorp/go-tfe"
	"github.com/mitchellh/iochan"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const helmTestValuesPath = "./charts/otfd/test-values.yaml"

// KubeDeploy is a deployment of OTF to a local kubernetes kind cluster. For
// testing purposes only.
type KubeDeploy struct {
	*tfe.Client
	KubeDeployConfig
	*debugger

	configPath string
	clientset  *kubernetes.Clientset
	tunnel     *exec.Cmd
	browser    *exec.Cmd
	siteToken  string
}

type KubeDeployConfig struct {
	Namespace          string
	RepoDir            string
	Release            string
	JobTTL             time.Duration
	OpenBrowser        bool
	CacheVolumeEnabled bool
}

func NewKubeDeploy(ctx context.Context, cfg KubeDeployConfig) (*KubeDeploy, debugFunc, error) {
	if cfg.Namespace == "" {
		cfg.Namespace = petname.Generate(2, "-")
	}
	if cfg.Release == "" {
		cfg.Release = "otfd"
	}

	// Build and load docker images into kind
	cmd := exec.CommandContext(ctx, "make", "load")
	cmd.Dir = cfg.RepoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nodebug, fmt.Errorf("building and loading images: %w: %s", err, string(out))
	}

	// Write kind's kubeconfig to a temp file
	config, err := os.CreateTemp("", "otf-k8s-deploy-*")
	if err != nil {
		return nil, nodebug, err
	}

	cmd = exec.CommandContext(ctx, "kind", "get", "kubeconfig")
	var buf bytes.Buffer
	cmd.Stdout = config
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return nil, nodebug, fmt.Errorf("retrieving kind config: %w: %s", err, buf.String())
	}

	debugger := debugger{
		configPath: config.Name(),
		namespace:  cfg.Namespace,
	}

	// Install otfd helm chart
	args := []string{
		"upgrade",
		cfg.Release,
		filepath.Join(cfg.RepoDir, "charts/otfd"),
		"--install",
		"--kubeconfig", config.Name(),
		"--create-namespace",
		"--namespace", cfg.Namespace,
		"--values", filepath.Join(cfg.RepoDir, helmTestValuesPath),
		"--set", "image.tag=edge",
		"--set", "logging.http=true",
		"--set", "logging.verbosity=9",
		"--set", "runner.executor=kubernetes",
		"--set", "runner.pluginCache=true",
		"--set", "defaultEngine=tofu",
		"--wait",
		"--timeout", "60s",
		"--debug",
	}
	if cfg.JobTTL != 0 {
		args = append(args, "--set", fmt.Sprintf("runner.kubernetesTTLAfterFinish=%ds", cfg.JobTTL))
	}
	if cfg.CacheVolumeEnabled {
		args = append(args, "--set", "runner.cacheVolume.enabled=true")
	}
	cmd = exec.CommandContext(ctx, "helm", args...)
	out, err = cmd.CombinedOutput()
	if err != nil {

		return nil, debugger.debug("otfd"), fmt.Errorf("installing otfd helm chart: %w: %s", err, string(out))
	}

	// Create tunnel to otfd service so that we can communicate with it.
	r, w := io.Pipe()
	tunnel := exec.CommandContext(ctx,
		"kubectl",
		"-n", cfg.Namespace,
		"--kubeconfig", config.Name(),
		"port-forward",
		"services/otfd",
		":80",
	)
	tunnel.Stderr = &buf
	tunnel.Stdout = w
	if err := tunnel.Start(); err != nil {
		return nil, debugger.debug("otfd"), fmt.Errorf("creating tunnel to cluster: %w: %s", err, buf.String())
	}
	// Outputs something like:
	//Forwarding from 127.0.0.1:40009 -> 8080
	//Forwarding from [::1]:40009 -> 8080
	//
	// Grab random local listening port
	localPortRe := regexp.MustCompile(`Forwarding from 127.0.0.1:(\d+) -> 8080`)
	matches := localPortRe.FindStringSubmatch(<-iochan.DelimReader(r, '\n'))
	if matches == nil {
		return nil, debugger.debug("otfd"), fmt.Errorf("listening port not found in tunnel output: %s", buf.String())
	}
	localPort := matches[1]

	var browser *exec.Cmd
	if cfg.OpenBrowser {
		u := url.URL{Scheme: "http", Host: "localhost:" + localPort, Path: "/app/organizations"}
		browser = exec.CommandContext(ctx, "xdg-open", u.String())
		if err := browser.Start(); err != nil {
			return nil, nodebug, fmt.Errorf("opening browser to connect to local tunneled endpoint: %w", err)
		}
	}

	// Extract site token from helm test values file
	rawTestValues, err := os.ReadFile(filepath.Join(cfg.RepoDir, helmTestValuesPath))
	if err != nil {
		return nil, nodebug, fmt.Errorf("reading test values file: %w", err)
	}
	var testValues struct {
		SiteToken string `json:"siteToken"`
	}
	err = yaml.Unmarshal(rawTestValues, &testValues)
	if err != nil {
		return nil, nodebug, err
	}

	// Create TFE Client to talk to the remote otfd. We would use the OTF client
	// but not all endpoints are implemented.
	client, err := tfe.NewClient(&tfe.Config{
		Address: fmt.Sprintf("http://localhost:%s", localPort),
		Token:   testValues.SiteToken,
	})
	if err != nil {
		return nil, nodebug, err
	}

	// Build k8s client
	restcfg, err := clientcmd.BuildConfigFromFlags("", config.Name())
	if err != nil {
		return nil, debugger.debug("otfd"), err
	}
	clientset, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		return nil, nodebug, err
	}

	deploy := &KubeDeploy{
		KubeDeployConfig: cfg,
		configPath:       config.Name(),
		clientset:        clientset,
		tunnel:           tunnel,
		browser:          browser,
		Client:           client,
		siteToken:        testValues.SiteToken,
		debugger:         &debugger,
	}
	return deploy, nodebug, nil
}

func (k *KubeDeploy) Wait() error {
	return k.tunnel.Wait()
}

func (k *KubeDeploy) Close(deleteNamespace bool) error {
	if k.browser != nil {
		if err := k.browser.Process.Kill(); err != nil {
			return err
		}
	}
	if err := k.tunnel.Process.Kill(); err != nil {
		return err
	}
	if deleteNamespace {
		err := k.clientset.CoreV1().Namespaces().Delete(context.Background(), k.Namespace, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	if err := os.Remove(k.configPath); err != nil {
		return err
	}
	return nil
}

func (k *KubeDeploy) WaitPodSucceed(ctx context.Context, runID string, timeout time.Duration) (debugFunc, error) {
	if timeout == 0 {
		timeout = time.Second * 30
	}
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("otf.ninja/run-id=%s", runID),
	}
	err := k8swait.PollUntilContextTimeout(ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		pods, err := k.clientset.CoreV1().Pods(k.Namespace).List(ctx, opts)
		if err != nil {
			return false, err
		}
		if len(pods.Items) == 0 {
			return false, nil
		}
		switch pod := pods.Items[0]; pod.Status.Phase {
		case corev1.PodFailed:
			return false, errors.New("pod failed")
		case corev1.PodSucceeded:
			return true, nil
		}
		return false, nil
	})
	return k.debug("otf-job"), err
}

func (k *KubeDeploy) WaitJobAndSecretDeleted(ctx context.Context, runID string) error {
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("otf.ninja/run-id=%s", runID),
	}
	err := k8swait.PollUntilContextTimeout(
		ctx,
		time.Second,
		time.Second*30,
		true,
		func(ctx context.Context) (bool, error) {
			secrets, err := k.clientset.CoreV1().Secrets(k.Namespace).List(ctx, opts)
			if err != nil {
				return false, err
			}
			return len(secrets.Items) == 0, nil
		})
	if err != nil {
		return err
	}
	err = k8swait.PollUntilContextTimeout(
		ctx,
		time.Second,
		time.Second*30,
		true,
		func(ctx context.Context) (bool, error) {
			jobs, err := k.clientset.BatchV1().Jobs(k.Namespace).List(ctx, opts)
			if err != nil {
				return false, err
			}
			return len(jobs.Items) == 0, nil
		})
	return err
}

func (k *KubeDeploy) InstallAgentChart(ctx context.Context, token string) (debugFunc, error) {
	svc, err := k.clientset.CoreV1().Services(k.Namespace).Get(ctx, "otfd", metav1.GetOptions{})
	if err != nil {
		return k.debug("otfd"), err
	}
	// Install otf-agent helm chart
	args := []string{
		"install",
		"otf-agent",
		filepath.Join(k.RepoDir, "charts/otf-agent"),
		"--kubeconfig", k.configPath,
		"--namespace", k.Namespace,
		"--set", "image.tag=edge",
		// agent talks to otfd via its service ip
		"--set", "url=http://" + svc.Spec.ClusterIP,
		"--set", "token=" + token,
		"--set", "logging.http=true",
		"--set", "logging.verbosity=9",
		"--set", "runner.executor=kubernetes",
		"--wait",
		"--timeout", "60s",
		"--debug",
	}
	if k.JobTTL != 0 {
		args = append(args, "--set", "runner.kubernetesTTLAfterFinish=1s")
	}
	if k.CacheVolumeEnabled {
		args = append(args, "--set", "runner.cacheVolume.enabled=true")
	}
	cmd := exec.CommandContext(ctx, "helm", args...)
	return k.debug("otf-agent"), cmd.Run()
}

type debugger struct {
	configPath string
	namespace  string
}

type debugFunc func(context.Context) string

func (d *debugger) debug(component string) debugFunc {
	return func(ctx context.Context) string {
		var sb strings.Builder

		sb.WriteString("--- describe pods output ---\n")
		cmd := exec.CommandContext(ctx,
			"kubectl",
			"-n", d.namespace,
			"--kubeconfig", d.configPath,
			"describe",
			"pod",
		)
		describeOutput, err := cmd.CombinedOutput()
		if err != nil {
			sb.WriteString("error running kubectl describe pod: " + err.Error())
		} else {
			sb.Write(describeOutput)
		}
		sb.WriteString("\n\n")

		sb.WriteString("--- pod logs output ---\n")
		cmd = exec.CommandContext(ctx,
			"kubectl",
			"-n", d.namespace,
			"--kubeconfig", d.configPath,
			"logs",
			"-l", "app.kubernetes.io/name="+component,
		)
		logsOutput, err := cmd.CombinedOutput()
		if err != nil {
			sb.WriteString("error running kubectl logs: " + err.Error())
		} else {
			sb.Write(logsOutput)
		}
		sb.WriteString("\n\n")

		return sb.String()
	}
}

func nodebug(ctx context.Context) string { return "" }
