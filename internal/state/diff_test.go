package state

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff_NilFrom(t *testing.T) {
	to := &File{
		Resources: []Resource{
			{Name: "web", Type: "aws_instance", ProviderURI: `provider["registry.terraform.io/hashicorp/aws"]`},
		},
		Outputs: map[string]FileOutput{
			"url": {Value: json.RawMessage(`"https://example.com"`)},
		},
	}
	d := Diff(nil, to)
	assert.Len(t, d.Resources, 1)
	assert.Equal(t, ActionAdd, d.Resources[0].Action)
	assert.Equal(t, "web", d.Resources[0].Resource.Name)
	assert.Len(t, d.Outputs, 1)
	assert.Equal(t, ActionAdd, d.Outputs[0].Action)
	assert.Equal(t, "url", d.Outputs[0].Name)
}

func TestDiff_AddRemove(t *testing.T) {
	from := &File{
		Resources: []Resource{
			{Name: "old_bucket", Type: "aws_s3_bucket"},
		},
		Outputs: map[string]FileOutput{
			"gone": {Value: json.RawMessage(`"bye"`)},
		},
	}
	to := &File{
		Resources: []Resource{
			{Name: "new_instance", Type: "aws_instance"},
		},
		Outputs: map[string]FileOutput{
			"hello": {Value: json.RawMessage(`"hi"`)},
		},
	}
	d := Diff(from, to)

	if assert.Len(t, d.Resources, 2) {
		actions := map[string]ChangeAction{}
		for _, rc := range d.Resources {
			actions[rc.Resource.Name] = rc.Action
		}
		assert.Equal(t, ActionAdd, actions["new_instance"])
		assert.Equal(t, ActionRemove, actions["old_bucket"])
	}

	if assert.Len(t, d.Outputs, 2) {
		actions := map[string]ChangeAction{}
		for _, oc := range d.Outputs {
			actions[oc.Name] = oc.Action
		}
		assert.Equal(t, ActionAdd, actions["hello"])
		assert.Equal(t, ActionRemove, actions["gone"])
	}
}

func TestDiff_OutputChanged(t *testing.T) {
	from := &File{
		Outputs: map[string]FileOutput{
			"count": {Value: json.RawMessage(`3`)},
		},
	}
	to := &File{
		Outputs: map[string]FileOutput{
			"count": {Value: json.RawMessage(`7`)},
		},
	}
	d := Diff(from, to)
	assert.Empty(t, d.Resources)
	if assert.Len(t, d.Outputs, 1) {
		assert.Equal(t, ActionChange, d.Outputs[0].Action)
		assert.Equal(t, "count", d.Outputs[0].Name)
	}
}

func TestDiff_Unchanged(t *testing.T) {
	f := &File{
		Resources: []Resource{{Name: "web", Type: "aws_instance"}},
		Outputs:   map[string]FileOutput{"x": {Value: json.RawMessage(`1`)}},
	}
	d := Diff(f, f)
	assert.False(t, d.HasChanges())
}
