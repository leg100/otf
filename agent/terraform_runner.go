package agent

import (
	"bytes"
	"context"
	"os/exec"
)

type TerraformRunner interface {
	Plan(context.Context, string) ([]byte, error)
}

type runner struct{}

func (r *runner) Plan(ctx context.Context, path string) ([]byte, error) {
	initOut, err := r.run(ctx, "init", path)
	if err != nil {
		return nil, err
	}

	planOut, err := r.run(ctx, "plan", path)
	if err != nil {
		return nil, err
	}

	return append(initOut, planOut...), nil
}

func (r *runner) run(ctx context.Context, command, path string) ([]byte, error) {
	buf := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, "terraform", command, "-no-color")
	cmd.Dir = path
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
