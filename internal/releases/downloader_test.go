package releases

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

	"github.com/leg100/otf/internal/engine"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEngine struct {
	engine.Engine
	u *url.URL
}

func (e *testEngine) SourceURL(version string) *url.URL { return e.u }
func (e *testEngine) String() string                    { return "terraform" }

func TestDownloader(t *testing.T) {
	// setup web server
	http.Handle("/", http.FileServer(http.Dir("testdata")))
	srv := httptest.NewTLSServer(nil)
	t.Cleanup(func() {
		srv.Close()
	})
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	u.Path = fmt.Sprintf("/terraform/1.2.3/terraform_1.2.3_%s_%s.zip", runtime.GOOS, runtime.GOARCH)

	engine := &testEngine{
		u: u,
	}

	dl := NewDownloader(engine, t.TempDir())
	dl.client = &http.Client{Transport: otfhttp.InsecureTransport}

	buf := new(bytes.Buffer)
	tfpath, err := dl.Download(context.Background(), "1.2.3", buf)
	require.NoError(t, err)
	require.FileExists(t, tfpath)
	tfbin, err := os.ReadFile(tfpath)
	require.NoError(t, err)
	assert.Equal(t, "I am a fake terraform binary\n", string(tfbin))
	assert.Equal(t, "downloading terraform, version 1.2.3\n", buf.String())
}
