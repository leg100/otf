package engine

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"testing"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloader(t *testing.T) {
	// setup web server
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("testdata")))
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(func() {
		srv.Close()
	})
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	u.Path = fmt.Sprintf("/terraform/1.2.3/terraform_1.2.3_%s_%s.zip", runtime.GOOS, runtime.GOARCH)

	engine := &testEngine{
		u: u,
	}

	dl, err := NewDownloader(logr.Discard(), engine, t.TempDir())
	require.NoError(t, err)
	dl.client = &http.Client{Transport: otfhttp.InsecureTransport}

	// Download bin from fake server
	buf := new(bytes.Buffer)
	tfpath, err := dl.Download(context.Background(), "1.2.3", buf)
	require.NoError(t, err)
	require.FileExists(t, tfpath)
	tfbin, err := os.ReadFile(tfpath)
	require.NoError(t, err)
	assert.Equal(t, "I am a fake terraform binary\n", string(tfbin))
	assert.Equal(t, "downloading terraform, version 1.2.3\n", buf.String())

	// Request bin again. This time it should skip download.
	buf = new(bytes.Buffer)
	tfpath, err = dl.Download(context.Background(), "1.2.3", buf)
	require.NoError(t, err)
	require.FileExists(t, tfpath)
	tfbin, err = os.ReadFile(tfpath)
	require.NoError(t, err)
	assert.Equal(t, "I am a fake terraform binary\n", string(tfbin))
	assert.Equal(t, "", buf.String())
}
