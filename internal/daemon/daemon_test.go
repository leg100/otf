package daemon

import (
	"errors"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/require"
)

func TestDaemon_MissingSecretError(t *testing.T) {
	var missing *internal.ErrMissingParameter
	_, err := New(t.Context(), logr.Discard(), Config{})
	require.True(t, errors.As(err, &missing))
}
