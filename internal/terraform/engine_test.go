package terraform

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getLatestVersion(t *testing.T) {
	// endpoint is a stub endpoint that always returns 1.6.1 as latest
	// version
	endpoint := func() string {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			w.Write(testutils.ReadFile(t, "./testdata/latest.json"))
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)
		u, err := url.Parse(srv.URL)
		require.NoError(t, err)
		return u.String()
	}()

	got, err := getLatestVersion(endpoint)
	require.NoError(t, err)
	assert.Equal(t, "1.6.1", got)
}
