package variable

import (
	"context"
)

type fakeService struct {
	v           *Variable
	workspaceID string

	Service
}

func (f *fakeService) UpdateWorkspaceVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*Variable, string, error) {
	if err := f.v.update(nil, opts); err != nil {
		return nil, "", err
	}
	return f.v, f.workspaceID, nil
}
