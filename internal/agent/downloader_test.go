package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

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

	pathFinder := newTerraformPathFinder(t.TempDir())
	dl := newTerraformDownloader(pathFinder)
	dl.host = u.Host
	dl.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	buf := new(bytes.Buffer)
	tfpath, err := dl.download(context.Background(), "1.2.3", buf)
	require.NoError(t, err)
	require.FileExists(t, tfpath)
	tfbin, err := os.ReadFile(tfpath)
	require.NoError(t, err)
	assert.Equal(t, "I am a fake terraform binary\n", string(tfbin))
	assert.Equal(t, "downloading terraform, version 1.2.3\n", buf.String())
}
