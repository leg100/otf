package authenticator

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"

	otfapi "github.com/leg100/otf/internal/api"

	"github.com/gorilla/mux"
)

var (
	//go:embed icons/*
	icons embed.FS
)

func getIcon(name string) ([]byte, error) {
	// Default to openid unless github or gitlab
	switch name {
	case "github", "gitlab":
	default:
		name = "openid"
	}
	path := fmt.Sprintf("icons/%s.png", name)
	return icons.ReadFile(path)
}

type (
	api struct {
		*service
	}

	loginClient struct {
		Name        string `json:"name"`
		RequestPath string `json:"request_path"`
		Icon        []byte `json:"icon"`
	}
)

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.V2BasePath).Subrouter()
	r.HandleFunc("/login/clients", a.listLoginClients).Methods("GET")
}

func (a *api) listLoginClients(w http.ResponseWriter, r *http.Request) {
	clients := make([]loginClient, len(a.clients))
	for i, client := range a.clients {
		icon, err := getIcon(client.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		clients[i] = loginClient{
			Name:        client.Name,
			Icon:        icon,
			RequestPath: client.RequestPath(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")

	if err := json.NewEncoder(w).Encode(clients); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
