package state

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHCLType(t *testing.T) {
	tests := []struct {
		want  string
		value string
	}{
		{"bool", `true`},
		{"bool", `false`},
		{"number", `0.339`},
		{"number", `42`},
		{"string", `"item"`},
		{"tuple", `["item1", "item2"]`},
		{"object", `{"key1": "value1", "key2": "value2"}`},
		{"null", `null`},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got, err := newHCLType(json.RawMessage([]byte(tt.value)))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
