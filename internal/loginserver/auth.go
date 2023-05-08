package loginserver

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
)

func (s *server) authHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ClientID            string `schema:"client_id"`
		CodeChallenge       string `schema:"code_challenge"`
		CodeChallengeMethod string `schema:"code_challenge_method"`
		RedirectURI         string `schema:"redirect_uri"`
		ResponseType        string `schema:"response_type"`
		State               string `schema:"state"`

		Consented bool `schema:"consented"`
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
	re := redirectError{
		redirect: redirect,
		state:    params.State,
	}

	if params.ResponseType != "code" {
		re.error(w, r, ErrUnsupportedResponseType, "unsupported response type")
		return
	}

	if params.CodeChallengeMethod != "S256" {
		re.error(w, r, ErrInvalidRequest, "unsupported code challenge method")
		return
	}

	if r.Method == "GET" {
		s.Render("consent.tmpl", w, html.NewSitePage(r, "consent"))
		return
	}

	if !params.Consented {
		re.error(w, r, ErrAccessDenied, "user denied consent")
		return
	}

	user, err := auth.UserFromContext(r.Context())
	if err != nil {
		re.error(w, r, ErrServerError, err.Error())
		return
	}

	marshaled, err := json.Marshal(&authcode{
		CodeChallenge:       params.CodeChallenge,
		CodeChallengeMethod: params.CodeChallengeMethod,
		Username:            user.Username,
	})
	if err != nil {
		re.error(w, r, ErrServerError, err.Error())
		return
	}

	encrypted, err := internal.Encrypt(marshaled, s.secret)
	if err != nil {
		re.error(w, r, ErrServerError, err.Error())
		return
	}

	q := redirect.Query()
	q.Add("state", params.State)
	q.Add("code", encrypted)
	redirect.RawQuery = q.Encode()
	http.Redirect(w, r, redirect.String(), http.StatusFound)
}
