package html

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	renderer, err := NewRenderer(false)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = renderer.RenderTemplate("organization_new.tmpl", &buf, nil)
	require.NoError(t, err)
}
