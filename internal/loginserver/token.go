package loginserver

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/decode"
)

func (s *server) tokenHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ClientID     string `schema:"client_id"`
		Code         string `schema:"code"`
		CodeVerifier string `schema:"code_verifier"`
		GrantType    string `schema:"grant_type"`
		RedirectURI  string `schema:"redirect_uri"`
	}
	if err := decode.All(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	redirect, err := url.Parse(params.RedirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}

	if params.ClientID != ClientID {
		http.Error(w, ErrInvalidClient, http.StatusBadRequest)
		return
	}

	// errors from hereon in are sent to the redirect URI as per RFC6749.
	re := redirectError{redirect: redirect}

	if params.Code == "" {
		re.error(w, r, ErrInvalidRequest, "missing code")
		return
	}

	if params.CodeVerifier == "" {
		re.error(w, r, ErrInvalidRequest, "missing code verifier")
		return
	}

	if params.GrantType != "authorization_code" {
		re.error(w, r, ErrUnsupportedGrantType, "")
		return
	}

	decrypted, err := internal.Decrypt(params.Code, s.secret)
	if err != nil {
		re.error(w, r, ErrInvalidRequest, "decrypting authentication code: "+err.Error())
		return
	}

	var code authcode
	if err := json.Unmarshal(decrypted, &code); err != nil {
		re.error(w, r, ErrInvalidRequest, "unmarshaling authentication code: "+err.Error())
		return
	}

	// Perform PKCE authentication
	hash := sha256.Sum256([]byte(params.CodeVerifier))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:])
	if encoded != code.CodeChallenge {
		re.error(w, r, ErrInvalidGrant, encoded)
		return
	}

	// Create API token for user and include in response
	userCtx := internal.AddSubjectToContext(r.Context(), &auth.User{Username: code.Username})
	_, token, err := s.CreateUserToken(userCtx, auth.CreateUserTokenOptions{
		Description: "terraform login",
	})
	if err != nil {
		re.error(w, r, ErrInvalidRequest, err.Error())
		return
	}
	marshaled, err := json.Marshal(struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}{
		AccessToken: string(token),
		TokenType:   "bearer",
	})
	if err != nil {
		re.error(w, r, ErrInvalidRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(marshaled)
}
