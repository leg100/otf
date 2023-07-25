package variable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeService struct {
	variable *Variable

	Service
}

func newTestVariable(t *testing.T, workspaceID string, opts CreateVariableOptions) *Variable {
	v, err := fakeFactory().new(workspaceID, opts)
	require.NoError(t, err)
	return v
}

func fakeFactory() *factory {
	return &factory{
		generateVersion: func() string { return "" },
	}
}

func (f *fakeService) GetVariable(ctx context.Context, variableID string) (*Variable, error) {
	return f.variable, nil
}

func (f *fakeService) UpdateVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*Variable, error) {
	if err := fakeFactory().update(f.variable, opts); err != nil {
		return nil, err
	}
	return f.variable, nil
}
