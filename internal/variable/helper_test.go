package variable

import (
	"context"
)

type fakeService struct {
	v           *Variable
	workspaceID string

	Service
}

func (f *fakeService) UpdateWorkspaceVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*WorkspaceVariable, error) {
	if err := f.v.update(nil, opts, key); err != nil {
		return nil, err
	}
	return &WorkspaceVariable{WorkspaceID: f.workspaceID, Variable: f.v}, nil
}
