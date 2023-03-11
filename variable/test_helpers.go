package variable

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestVariable(t *testing.T, ws workspace.Workspace, opts CreateVariableOptions) *Variable {
	v, err := NewVariable(ws.ID, opts)
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
