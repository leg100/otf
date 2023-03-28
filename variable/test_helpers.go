package variable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestVariable(t *testing.T, workspaceID string, opts CreateVariableOptions) *Variable {
	v, err := NewVariable(workspaceID, opts)
	require.NoError(t, err)
	return v
}

type fakeService struct {
	variable *Variable

	Service
}

func (f *fakeService) GetVariable(ctx context.Context, variableID string) (*Variable, error) {
	return f.variable, nil
}

func (f *fakeService) UpdateVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*Variable, error) {
	if err := f.variable.Update(opts); err != nil {
		return nil, err
	}
	return f.variable, nil
}
