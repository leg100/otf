package variable

import (
	"context"

	"github.com/leg100/otf/internal/resource"
)

type fakeService struct {
	v           *Variable
	workspaceID resource.ID

	Service
}

func (f *fakeService) UpdateWorkspaceVariable(ctx context.Context, variableID resource.ID, opts UpdateVariableOptions) (*WorkspaceVariable, error) {
	if err := f.v.update(nil, opts); err != nil {
		return nil, err
	}
	return &WorkspaceVariable{WorkspaceID: f.workspaceID, Variable: f.v}, nil
}
