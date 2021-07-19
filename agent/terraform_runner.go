package agent

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type TerraformRunner interface {
	Plan(context.Context, string) ([]byte, []byte, []byte, error)
	Apply(context.Context, string) ([]byte, error)
}

type runner struct{}

func (r *runner) Plan(ctx context.Context, path string) ([]byte, []byte, []byte, error) {
	initOut, err := r.run(ctx, path, "init")
	if err != nil {
		return initOut, nil, nil, fmt.Errorf("terraform init failed: %w", err)
	}

	out, err := r.run(ctx, path, "plan", fmt.Sprintf("-out=%s", PlanFilename))
	if err != nil {
		return out, nil, nil, fmt.Errorf("terraform plan failed: %w", err)
	}

	json, err := r.run(ctx, path, "show", "-json", PlanFilename)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("terraform show failed: %w", err)
	}

	file, err := os.ReadFile(filepath.Join(path, PlanFilename))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to read plan file: %w", err)
	}

	return append(initOut, out...), file, json, nil
}

func (r *runner) Apply(ctx context.Context, path string) ([]byte, error) {
	initOut, err := r.run(ctx, path, "init")
	if err != nil {
		return initOut, fmt.Errorf("terraform init failed: %w", err)
	}

	out, err := r.run(ctx, path, "apply", PlanFilename)
	if err != nil {
		return out, fmt.Errorf("terraform apply failed: %w", err)
	}

	return append(initOut, out...), nil
}

func (r *runner) run(ctx context.Context, path, command string, args ...string) ([]byte, error) {
	buf := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, "terraform", append([]string{command, "-no-color"}, args...)...)
	cmd.Dir = path
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}
