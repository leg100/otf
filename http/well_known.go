package http

import (
	"encoding/json"
	"net/http"
)

var paths = WellKnown{
	ModulesV1:  "/api/registry/v1/modules/",
	MotdV1:     "/api/terraform/motd",
	StateV2:    "/api/v2/",
	TfeV2:      "/api/v2/",
	TfeV21:     "/api/v2/",
	TfeV22:     "/api/v2/",
	VersionsV1: "https://checkpoint-api.hashicorp.com/v1/versions/",
}

type WellKnown struct {
	ModulesV1  string `json:"modules.v1"`
	MotdV1     string `json:"motd.v1"`
	StateV2    string `json:"state.v2"`
	TfeV2      string `json:"tfe.v2"`
	TfeV21     string `json:"tfe.v2.1"`
	TfeV22     string `json:"tfe.v2.2"`
	VersionsV1 string `json:"versions.v1"`
}

func (s *Server) WellKnown(w http.ResponseWriter, r *http.Request) {
	payload, err := json.Marshal(paths)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	w.Header().Set("Content-type", jsonApplication)
	w.Write(payload)
}
