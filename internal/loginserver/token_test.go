package loginserver

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/require"
)

func TestTokenHandler(t *testing.T) {
	secret := internal.GenerateRandomString(32)
	srv := fakeServer(t, secret)

	verifier := "myverifier"
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	mashaled, err := json.Marshal(&authcode{
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		Username:            "bobby",
	})
	require.NoError(t, err)
	code, err := internal.Encrypt(mashaled, []byte(secret))
	require.NoError(t, err)

	q := "/?"
	q += "redirect_uri=https://localhost:10000"
	q += "&client_id=terraform"
	q += "&grant_type=authorization_code"
	q += "&code=" + code
	q += "&code_verifier=" + verifier

	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	srv.tokenHandler(w, r)

	require.Equal(t, 200, w.Code, w.Body.String())

	//decrypted, err := internal.Decrypt(w.Body.String(), secret)
	//require.NoError(t, err)

	var response struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
}
