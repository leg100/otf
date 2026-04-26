package session

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_StartSession(t *testing.T) {
	svc := Service{
		logger: logr.Discard(),
		client: &fakeSessionClient{},
	}

	userID := resource.NewTfeID(resource.UserKind)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?", nil)
	err := svc.StartSession(w, r, userID)
	require.NoError(t, err)

	// verify and validate token in cookie set in response
	cookies := w.Result().Cookies()
	require.Equal(t, 1, len(cookies))
	assert.Equal(t, "fake token", cookies[0].Value)

	// user is redirected to their profile page
	assert.Equal(t, 302, w.Code)
	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, path.Profile(), loc.Path)
}

type fakeSessionClient struct{}

func (f *fakeSessionClient) NewToken(subjectID resource.TfeID, expiry *time.Time) ([]byte, error) {
	return []byte("fake token"), nil
}
