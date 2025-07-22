package gitlab

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/require"
)

// setup sets up a test HTTP server along with a gitlab.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup(t *testing.T) (*http.ServeMux, *Client) {
	// mux is the HTTP request multiplexer used with the test server.
	mux := http.NewServeMux()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	// client is the Gitlab client being tested.
	client, err := NewClient(ClientOptions{
		BaseURL:             &internal.WebURL{URL: *u},
		SkipTLSVerification: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return mux, client
}
