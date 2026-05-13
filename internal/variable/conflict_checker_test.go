package variable

import (
	"context"
	"errors"
	"testing"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeConflictClient is a test double for conflictCheckerClient.
type fakeConflictClient struct {
	set             *VariableSet
	getSetErr       error
	variables       []*Variable
	listVarsErr     error
	globalVariables []*Variable
	listGlobalErr   error
}

func (f *fakeConflictClient) getVariableSet(_ context.Context, _ resource.ID) (*VariableSet, error) {
	return f.set, f.getSetErr
}

func (f *fakeConflictClient) listVariables(_ context.Context, _ resource.TfeID) ([]*Variable, error) {
	return f.variables, f.listVarsErr
}

func (f *fakeConflictClient) listGlobalVariables(_ context.Context, _ organization.Name) ([]*Variable, error) {
	return f.globalVariables, f.listGlobalErr
}

// makeVar builds a minimal variable with the given ID (zero means generate new), key, and category.
func makeVar(id resource.TfeID, key string, cat VariableCategory) *Variable {
	if id.IsZero() {
		id = resource.NewTfeID(resource.VariableKind)
	}
	return &Variable{ID: id, Key: key, Category: cat}
}

func TestConflictChecker_checkVariable(t *testing.T) {
	workspaceID := resource.NewTfeID(resource.WorkspaceKind)
	setID := resource.NewTfeID(resource.VariableSetKind)
	orgName, err := organization.NewName("my-org")
	require.NoError(t, err)

	sharedID := resource.NewTfeID(resource.VariableKind)
	clientErr := errors.New("client error")

	tests := []struct {
		name     string
		variable *Variable
		client   *fakeConflictClient
		wantErr  error
	}{
		{
			name: "workspace variable - no conflict",
			variable: &Variable{
				ID:       resource.NewTfeID(resource.VariableKind),
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: workspaceID,
			},
			client: &fakeConflictClient{
				variables: []*Variable{
					makeVar(resource.TfeID{}, "bar", CategoryTerraform), // different key
					makeVar(resource.TfeID{}, "foo", CategoryEnv),       // same key, different category
				},
			},
			wantErr: nil,
		},
		{
			name: "workspace variable - conflict",
			variable: &Variable{
				ID:       resource.NewTfeID(resource.VariableKind),
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: workspaceID,
			},
			client: &fakeConflictClient{
				variables: []*Variable{makeVar(resource.TfeID{}, "foo", CategoryTerraform)},
			},
			wantErr: ErrVariableConflict,
		},
		{
			name: "workspace variable - no conflict with itself",
			variable: &Variable{
				ID:       sharedID,
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: workspaceID,
			},
			client: &fakeConflictClient{
				// same ID returned — must not be treated as a conflict
				variables: []*Variable{makeVar(sharedID, "foo", CategoryTerraform)},
			},
		},
		{
			name: "workspace variable - listVariables error",
			variable: &Variable{
				ID:       resource.NewTfeID(resource.VariableKind),
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: workspaceID,
			},
			client:  &fakeConflictClient{listVarsErr: clientErr},
			wantErr: clientErr,
		},
		{
			name: "non-global set variable - no conflict",
			variable: &Variable{
				ID:       resource.NewTfeID(resource.VariableKind),
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: setID,
			},
			client: &fakeConflictClient{
				set:       &VariableSet{ID: setID, Global: false, Organization: orgName},
				variables: []*Variable{makeVar(resource.TfeID{}, "bar", CategoryTerraform)},
			},
		},
		{
			name: "non-global set variable - conflict",
			variable: &Variable{
				ID:       resource.NewTfeID(resource.VariableKind),
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: setID,
			},
			client: &fakeConflictClient{
				set:       &VariableSet{ID: setID, Global: false, Organization: orgName},
				variables: []*Variable{makeVar(resource.TfeID{}, "foo", CategoryTerraform)},
			},
			wantErr: ErrVariableConflict,
		},
		{
			name: "global set variable - no conflict",
			variable: &Variable{
				ID:       resource.NewTfeID(resource.VariableKind),
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: setID,
			},
			client: &fakeConflictClient{
				set:             &VariableSet{ID: setID, Global: true, Organization: orgName},
				globalVariables: []*Variable{makeVar(resource.TfeID{}, "bar", CategoryTerraform)},
			},
		},
		{
			name: "global set variable - conflict with other global set",
			variable: &Variable{
				ID:       resource.NewTfeID(resource.VariableKind),
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: setID,
			},
			client: &fakeConflictClient{
				set:             &VariableSet{ID: setID, Global: true, Organization: orgName},
				globalVariables: []*Variable{makeVar(resource.TfeID{}, "foo", CategoryTerraform)},
			},
			wantErr: ErrVariableConflict,
		},
		{
			name: "global set variable - getVariableSet error",
			variable: &Variable{
				ID:       resource.NewTfeID(resource.VariableKind),
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: setID,
			},
			client:  &fakeConflictClient{getSetErr: clientErr},
			wantErr: clientErr,
		},
		{
			name: "global set variable - listGlobalVariables error",
			variable: &Variable{
				ID:       resource.NewTfeID(resource.VariableKind),
				Key:      "foo",
				Category: CategoryTerraform,
				ParentID: setID,
			},
			client: &fakeConflictClient{
				set:           &VariableSet{ID: setID, Global: true, Organization: orgName},
				listGlobalErr: clientErr,
			},
			wantErr: clientErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &conflictChecker{client: tt.client}
			err := c.checkVariable(context.Background(), tt.variable)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConflictChecker_checkSet(t *testing.T) {
	setID := resource.NewTfeID(resource.VariableSetKind)
	orgName, err := organization.NewName("my-org")
	require.NoError(t, err)

	sharedID := resource.NewTfeID(resource.VariableKind)
	dbErr := errors.New("db error")

	tests := []struct {
		name    string
		set     *VariableSet
		client  *fakeConflictClient
		wantErr error
	}{
		{
			name: "non-global set - skipped",
			set:  &VariableSet{ID: setID, Global: false, Organization: orgName},
			// listVarsErr set to prove the db is never called
			client: &fakeConflictClient{listVarsErr: errors.New("should not be called")},
		},
		{
			name: "global set - no conflict",
			set:  &VariableSet{ID: setID, Global: true, Organization: orgName},
			client: &fakeConflictClient{
				variables:       []*Variable{makeVar(resource.TfeID{}, "foo", CategoryTerraform)},
				globalVariables: []*Variable{makeVar(resource.TfeID{}, "bar", CategoryTerraform)},
			},
		},
		{
			name: "global set - conflict with another global set variable",
			set:  &VariableSet{ID: setID, Global: true, Organization: orgName},
			client: &fakeConflictClient{
				variables:       []*Variable{makeVar(resource.TfeID{}, "foo", CategoryTerraform)},
				globalVariables: []*Variable{makeVar(resource.TfeID{}, "foo", CategoryTerraform)},
			},
			wantErr: ErrVariableConflict,
		},
		{
			name: "global set - no conflict with itself",
			set:  &VariableSet{ID: setID, Global: true, Organization: orgName},
			client: &fakeConflictClient{
				// set's own variable appears in the global scope with the same ID
				variables:       []*Variable{makeVar(sharedID, "foo", CategoryTerraform)},
				globalVariables: []*Variable{makeVar(sharedID, "foo", CategoryTerraform)},
			},
		},
		{
			name:    "global set - listVariables error",
			set:     &VariableSet{ID: setID, Global: true, Organization: orgName},
			client:  &fakeConflictClient{listVarsErr: dbErr},
			wantErr: dbErr,
		},
		{
			name: "global set - listGlobalVariables error",
			set:  &VariableSet{ID: setID, Global: true, Organization: orgName},
			client: &fakeConflictClient{
				variables:     []*Variable{makeVar(resource.TfeID{}, "foo", CategoryTerraform)},
				listGlobalErr: dbErr,
			},
			wantErr: dbErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &conflictChecker{client: tt.client}
			err := c.checkSet(context.Background(), tt.set)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
