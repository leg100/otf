package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

type (
	v2 struct {
		*Service
		*tfeapi.Responder
	}
)

func (a *v2) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.V2BasePath).Subrouter()

	r.HandleFunc("/current-user", a.getCurrentUser).Methods("GET")
	r.HandleFunc("/logout", a.logout).Methods("POST")
}

func (a *v2) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	subject, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := struct {
		Username string `json:"username"`
	}{
		Username: subject.String(),
	}

	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// logout the user by simply purging the session cookie from the browser cookie
// jar.
func (a *v2) logout(w http.ResponseWriter, r *http.Request) {
	html.SetCookie(w, tokens.SessionCookie, "", &time.Time{})
}
