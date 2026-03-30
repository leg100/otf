package healthz

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

type fakeClient struct {
	err error
}

func (f *fakeClient) Ping(_ context.Context) error {
	return f.err
}

func newTestServer(client CheckClient) *httptest.Server {
	r := mux.NewRouter()
	c := &Check{Client: client}
	c.AddHandlers(r)
	return httptest.NewServer(r)
}

func TestHealthz_OK(t *testing.T) {
	srv := newTestServer(&fakeClient{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-type"); ct != "application/json" {
		t.Errorf("want Content-type application/json, got %q", ct)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, `"OK"`) {
		t.Errorf("want body to contain OK status, got %q", body)
	}
}

func TestHealthz_Unhealthy(t *testing.T) {
	pingErr := errors.New("connection refused")
	srv := newTestServer(&fakeClient{err: pingErr})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("want 503, got %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, `"UNHEALTHY"`) {
		t.Errorf("want body to contain UNHEALTHY status, got %q", body)
	}
	if !strings.Contains(body, pingErr.Error()) {
		t.Errorf("want body to contain error message %q, got %q", pingErr.Error(), body)
	}
}
