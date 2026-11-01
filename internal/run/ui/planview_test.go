package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildChangeView(t *testing.T) {
	planJSON := []byte(`{
	  "format_version": "1.2",
	  "resource_changes": [
	    {
	      "address": "a_resource",
	      "change": {
	        "actions": ["no-op"],
	        "before": {"id": "1"}, "after": {"id": "1"}
	      }
	    },
	    {
	      "address": "b_resource",
	      "change": {
	        "actions": ["create"],
	        "before": null,
	        "after": {"name": "b", "port": 443},
	        "after_sensitive": {"name": true}
	      }
	    },
	    {
	      "address": "c_resource",
	      "change": {
	        "actions": ["update"],
	        "before": {"port": 80, "keep": "x"},
	        "after": {"port": 443, "keep": "x", "fingerprint": null},
	        "after_unknown": {"port": false, "fingerprint": true}
	      }
	    },
	    {
	      "address": "d_resource",
	      "change": {
	        "actions": ["delete", "create"],
	        "before": {"password": "old"},
	        "after": {"password": "new"},
	        "before_sensitive": true
	      }
	    },
	    {
	      "address": "e_resource",
	      "change": {
	        "actions": ["delete"],
	        "before": {"name": "e"},
	        "after": null
	      }
	    }
	  ]
	}`)

	changes, err := buildChangeView(planJSON)
	require.NoError(t, err)

	// no-op changes are skipped; the rest are grouped: create, update,
	// replace, delete.
	require.Len(t, changes, 4)
	assert.Equal(t, []changeView{
		{
			Action:  "create",
			Address: "b_resource",
			Fields: []fieldView{
				// unchanged absent attributes are not listed; port is
				// listed with no before value; sensitive name is masked
				{Key: "name", Before: "(sensitive value)", After: "(sensitive value)"},
				{Key: "port", Before: "", After: "443"},
			},
		},
		{
			Action:  "update",
			Address: "c_resource",
			// only the changed attribute is listed; unknown is annotated;
			// unchanged keep is absent
			Fields: []fieldView{
				{Key: "fingerprint", Before: "", After: "(known after apply)"},
				{Key: "port", Before: "80", After: "443"},
			},
		},
		{
			Action:  "replace",
			Address: "d_resource",
			// whole-object sensitivity masks every attribute
			Fields: []fieldView{
				{Key: "password", Before: "(sensitive value)", After: "(sensitive value)"},
			},
		},
		{
			Action:  "delete",
			Address: "e_resource",
			Fields: []fieldView{
				// destroyed attributes show no after value
				{Key: "name", Before: "e", After: ""},
			},
		},
	}, changes)
}

func TestBuildChangeViewEmpty(t *testing.T) {
	// An empty plan yields a non-nil empty view so the caller can
	// distinguish "unavailable" from "available and empty".
	changes, err := buildChangeView([]byte(`{"resource_changes": []}`))
	require.NoError(t, err)
	assert.NotNil(t, changes)
	assert.Empty(t, changes)
}

func TestBuildChangeViewInvalid(t *testing.T) {
	_, err := buildChangeView([]byte(`{not json`))
	assert.Error(t, err)
}
