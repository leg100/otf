package releases

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloader(t *testing.T) {
	// setup web server
	http.Handle("/", http.FileServer(http.Dir("testdata/releases")))
	srv := httptest.NewTLSServer(nil)
	t.Cleanup(func() {
		srv.Close()
	})
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	dl := newDownloader(t.TempDir())
	dl.host = u.Host
	dl.client = &http.Client{
		Transport: otfhttp.DefaultTransport(true),
	}

	buf := new(bytes.Buffer)
	tfpath, err := dl.Download(context.Background(), "1.2.3", buf)
	require.NoError(t, err)
	require.FileExists(t, tfpath)
	tfbin, err := os.ReadFile(tfpath)
	require.NoError(t, err)
	assert.Equal(t, "I am a fake terraform binary\n", string(tfbin))
	assert.Equal(t, "downloading terraform, version 1.2.3\n", buf.String())
}
