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

	service
}

func (f *fakeService) update(ctx context.Context, variableID string, opts UpdateVariableOptions) (*Variable, error) {
	if err := f.variable.Update(opts); err != nil {
		return nil, err
	}
	return f.variable, nil
}
