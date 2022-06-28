package http

import (
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2/github"
)

var defaultWellKnownPaths = WellKnown{
	ModulesV1:  "/api/registry/v1/modules/",
	MotdV1:     "/api/terraform/motd",
	StateV2:    "/api/v2/",
	TfeV2:      "/api/v2/",
	TfeV21:     "/api/v2/",
	TfeV22:     "/api/v2/",
	VersionsV1: "https://checkpoint-api.hashicorp.com/v1/versions/",
}

type WellKnown struct {
	LoginV1    LoginService `json:"login.v1"`
	ModulesV1  string       `json:"modules.v1"`
	MotdV1     string       `json:"motd.v1"`
	StateV2    string       `json:"state.v2"`
	TfeV2      string       `json:"tfe.v2"`
	TfeV21     string       `json:"tfe.v2.1"`
	TfeV22     string       `json:"tfe.v2.2"`
	VersionsV1 string       `json:"versions.v1"`
}

type LoginService struct {
	Client     string   `json:"client"`
	GrantTypes []string `json:"grant_types"`
	Authz      string   `json:"authz"`
	Token      string   `json:"token"`
	Ports      []int    `json:"ports"`
}

func (s *Server) WellKnown(w http.ResponseWriter, r *http.Request) {
	defaultWellKnownPaths.LoginV1 = LoginService{
		Client:     s.ApplicationConfig.Github.ClientID,
		GrantTypes: []string{"authz_code"},
		Authz:      github.Endpoint.AuthURL,
		Token:      github.Endpoint.TokenURL,
		Ports:      []int{10000, 10010},
	}
	payload, err := json.Marshal(defaultWellKnownPaths)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	w.Header().Set("Content-type", jsonApplication)
	w.Write(payload)
}
