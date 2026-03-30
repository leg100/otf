package healthz

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeClient struct {
	err error
}

func (f *fakeClient) Ping(_ context.Context) error {
	return f.err
}

func TestHealthz_OK(t *testing.T) {
	check := &Check{Client: &fakeClient{}}

	assert.HTTPStatusCode(t, check.healthz, "GET", "/healthz", nil, 200)

	got := assert.HTTPBody(check.healthz, "GET", "/healthz", nil)
	assert.Equal(t, `{"status":"OK"}`, strings.TrimSpace(got))
}

func TestHealthz_Unhealthy(t *testing.T) {
	pingErr := errors.New("connection refused")
	check := &Check{Client: &fakeClient{err: pingErr}}

	assert.HTTPStatusCode(t, check.healthz, "GET", "/healthz", nil, 503)

	got := assert.HTTPBody(check.healthz, "GET", "/healthz", nil)
	assert.Equal(t, `{"status":"UNHEALTHY","error":"connection refused"}`, strings.TrimSpace(got))
}
