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
}

type runner struct{}

func (r *runner) Plan(ctx context.Context, path string) ([]byte, []byte, []byte, error) {
	initOut, err := r.run(ctx, path, "init", "-no-color")
	if err != nil {
		return nil, nil, nil, err
	}

	dir, err := os.MkdirTemp("", "plans")
	if err != nil {
		return nil, nil, nil, err
	}
	planPath := filepath.Join(dir, "plan.out")

	out, err := r.run(ctx, path, "plan", "-no-color", fmt.Sprintf("-out=%s", planPath))
	if err != nil {
		return nil, nil, nil, err
	}

	json, err := r.run(ctx, path, "show", "-no-color", "-json", planPath)
	if err != nil {
		return nil, nil, nil, err
	}

	file, err := os.ReadFile(planPath)
	if err != nil {
		return nil, nil, nil, err
	}

	return append(initOut, out...), file, json, nil
}

func (r *runner) run(ctx context.Context, path, command string, args ...string) ([]byte, error) {
	buf := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, "terraform", append([]string{command, "-no-color"}, args...)...)
	cmd.Dir = path
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
