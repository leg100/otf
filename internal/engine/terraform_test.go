package engine

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getLatestTerraformVersion(t *testing.T) {
	// endpoint is a stub endpoint that always returns 1.6.1 as latest
	// version
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write(testutils.ReadFile(t, "./testdata/terraform/latest.json"))
	}
	srv := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(srv.Close)

	getter := &terraformClient{
		endpoint: srv.URL,
	}

	got, err := getter.getLatestVersion(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "1.6.1", got)
}
