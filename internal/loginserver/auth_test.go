package loginserver

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_ConsentUI(t *testing.T) {
	srv := fakeServer(t, testutils.NewSecret(t))

	q := "/?"
	q += "redirect_uri=https://localhost:10000"
	q += "&client_id=terraform"
	q += "&response_type=code"

	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	srv.authHandler(w, r)
}

func TestAuthHandler_Auth(t *testing.T) {
	secret := testutils.NewSecret(t)
	srv := fakeServer(t, secret)

	q := "/?"
	q += "redirect_uri=https://localhost:10000"
	q += "&client_id=terraform"
	q += "&response_type=code"
	q += "&consented=true"
	q += "&code_challenge_method=S256"
	q += "&state=somethingrandom"

	r := httptest.NewRequest("POST", q, nil)
	r = r.WithContext(authz.AddSubjectToContext(r.Context(), &user.User{Username: "bobby"}))
	w := httptest.NewRecorder()
	srv.authHandler(w, r)

	// check redirect URI
	require.Equal(t, 302, w.Code)
	redirect, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "localhost:10000", redirect.Host)
	assert.Equal(t, "somethingrandom", redirect.Query().Get("state"))
	// ensure we haven't receive an oauth error payload
	require.Empty(t, redirect.Query().Get("error"))

	// check contents of auth code
	encrypted := redirect.Query().Get("code")
	decrypted, err := internal.Decrypt(encrypted, secret)
	require.NoError(t, err)
	var code authcode
	err = json.Unmarshal(decrypted, &code)
	require.NoError(t, err)
	assert.Equal(t, "bobby", code.Username)
}
