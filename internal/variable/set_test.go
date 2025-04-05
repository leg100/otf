package variable

import (
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestVariableSet_checkGlobalConflicts(t *testing.T) {
	organizationSets := []*VariableSet{
		{
			ID:     resource.NewTfeID(resource.VariableSetKind),
			Global: false,
			Variables: []*Variable{
				{
					ID:       resource.NewTfeID(resource.VariableKind),
					Key:      "foo",
					Category: CategoryTerraform,
				},
			},
		},
		{
			ID:     resource.NewTfeID(resource.VariableSetKind),
			Global: true,
			Variables: []*Variable{
				{
					ID:       resource.NewTfeID(resource.VariableKind),
					Key:      "foo",
					Category: CategoryTerraform,
				},
			},
		},
	}

	tests := []struct {
		name string
		set  VariableSet
		want error
	}{
		{
			name: "non-global set does not conflict",
			set:  VariableSet{},
		},
		{
			name: "conflict",
			set: VariableSet{
				Global: true,
				Variables: []*Variable{
					{
						Key:      "foo",
						Category: CategoryTerraform,
					},
				},
			},
			want: ErrVariableConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.set.checkGlobalConflicts(organizationSets)
			assert.Equal(t, tt.want, got)
		})
	}
}
