package http

import (
	"crypto/rand"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadConfigurationVersion(t *testing.T) {
	// fake server
	srv := &Server{
		Application: &fakeConfigurationVersionApp{},
		Logger:      logr.Discard(),
	}

	// setup web server
	router := mux.NewRouter()
	router.Handle("/upload/{id}", srv.UploadConfigurationVersion())
	webSrv := httptest.NewTLSServer(router)
	t.Cleanup(func() {
		webSrv.Close()
	})
	url := webSrv.URL + "/upload/cv-123"

	// setup client
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	t.Run("small upload", func(t *testing.T) {
		smallRandReader := io.LimitReader(rand.Reader, 100)
		req, err := http.NewRequest("PUT", url, smallRandReader)
		req.ContentLength = 100
		require.NoError(t, err)
		res, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
	})

	t.Run("excessively big upload", func(t *testing.T) {
		largeRandReader := io.LimitReader(rand.Reader, otf.ConfigMaxSize+1)
		req, err := http.NewRequest("PUT", url, largeRandReader)
		req.ContentLength = otf.ConfigMaxSize + 1
		require.NoError(t, err)
		res, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 422, res.StatusCode)
	})
}
