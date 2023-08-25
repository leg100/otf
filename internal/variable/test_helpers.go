package variable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeService struct {
	wv *WorkspaceVariable

	Service
}

func newTestVariable(t *testing.T, workspaceID string, opts CreateVariableOptions) *WorkspaceVariable {
	v, err := fakeFactory().new(opts)
	require.NoError(t, err)
	return &WorkspaceVariable{WorkspaceID: workspaceID, Variable: v}
}

func fakeFactory() *factory {
	return &factory{
		generateVersion: func() string { return "" },
	}
}

func (f *fakeService) GetWorkspaceVariable(ctx context.Context, variableID string) (*WorkspaceVariable, error) {
	return f.wv, nil
}

func (f *fakeService) UpdateWorkspaceVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*WorkspaceVariable, error) {
	if err := fakeFactory().update(f.wv.Variable, opts); err != nil {
		return nil, err
	}
	return f.wv, nil
}
