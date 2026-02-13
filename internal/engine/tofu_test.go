package engine

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/leg100/otf/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getLatestVersion(t *testing.T) {
	_, u := github.NewTestServer(t,
		github.WithHandler("/api/v3/repos/opentofu/opentofu/releases/latest", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			f, err := os.Open("./testdata/tofu/latest.json")
			require.NoError(t, err)
			io.Copy(w, f)
			f.Close()

		}),
		github.WithDisableTLS(),
	)
	got, err := getLatestTofuVersion(context.Background(), new(u.String()))
	require.NoError(t, err)
	assert.Equal(t, "1.9.0", got)
}
