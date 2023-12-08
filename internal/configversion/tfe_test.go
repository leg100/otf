package configversion

import (
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	otfhttp "github.com/leg100/otf/internal/http"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadConfigurationVersion(t *testing.T) {
	api := &tfe{
		// only permit upto 100 byte uploads
		maxConfigSize: 100,
		tfeClient:     &fakeConfigService{},
	}

	// setup web server
	router := mux.NewRouter()
	router.Handle("/upload/{id}", api.uploadConfigurationVersion())
	webSrv := httptest.NewTLSServer(router)
	t.Cleanup(webSrv.Close)
	url := webSrv.URL + "/upload/cv-123"

	// setup client
	client := http.Client{
		Transport: otfhttp.InsecureTransport,
	}

	// upload config smaller than MaxConfigSize
	t.Run("small upload", func(t *testing.T) {
		smallRandReader := io.LimitReader(rand.Reader, 99)
		req, err := http.NewRequest("PUT", url, smallRandReader)
		req.ContentLength = 99
		require.NoError(t, err)
		res, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
	})

	// upload config bigger than MaxConfigSize
	t.Run("excessively big upload", func(t *testing.T) {
		largeRandReader := io.LimitReader(rand.Reader, 101)
		req, err := http.NewRequest("PUT", url, largeRandReader)
		req.ContentLength = 101
		require.NoError(t, err)
		res, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 422, res.StatusCode)
	})
}
