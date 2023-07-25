package variable

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
)

func TestFactory_Update(t *testing.T) {
	tests := []struct {
		name   string
		opts   UpdateVariableOptions
		before Variable
		after  Variable
		err    bool // want error
	}{
		{
			name: "no change",
			opts: UpdateVariableOptions{},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
		},
		{
			name: "key",
			opts: UpdateVariableOptions{Key: internal.String("teddy")},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:      "teddy",
				Value:    "bar",
				Category: CategoryTerraform,
			},
		},
		{
			name: "value",
			opts: UpdateVariableOptions{Value: internal.String("baz")},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:      "foo",
				Value:    "baz",
				Category: CategoryTerraform,
			},
		},
		{
			name: "non-sensitive to sensitive",
			opts: UpdateVariableOptions{Sensitive: internal.Bool(true)},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:       "foo",
				Value:     "bar",
				Category:  CategoryTerraform,
				Sensitive: true,
			},
		},
		{
			name: "non-hcl to hcl",
			opts: UpdateVariableOptions{HCL: internal.Bool(true)},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
				HCL:      true,
			},
		},
		{
			name: "sensitive to non-sensitive",
			opts: UpdateVariableOptions{Sensitive: internal.Bool(false)},
			before: Variable{
				Key:       "foo",
				Value:     "bar",
				Category:  CategoryTerraform,
				Sensitive: true,
			},
			err: true,
		},
	}
	for _, tt := range tests {
		f := fakeFactory()
		t.Run(tt.name, func(t *testing.T) {
			got := tt.before
			err := f.update(&got, tt.opts)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.after, got)
			}
		})
	}
}
