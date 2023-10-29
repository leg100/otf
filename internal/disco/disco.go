// Package disco implements terraform's "remote service discovery protocol":
//
// https://developer.hashicorp.com/terraform/internals/v1.3.x/remote-service-discovery
package disco

import (
	gohttp "net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/json"
	"github.com/leg100/otf/internal/loginserver"
	"github.com/leg100/otf/internal/tfeapi"
)

var discoveryPayload = json.MustMarshal(struct {
	ModulesV1 string                    `json:"modules.v1"`
	MotdV1    string                    `json:"motd.v1"`
	StateV2   string                    `json:"state.v2"`
	TfeV2     string                    `json:"tfe.v2"`
	TfeV21    string                    `json:"tfe.v2.1"`
	TfeV22    string                    `json:"tfe.v2.2"`
	LoginV1   loginserver.DiscoverySpec `json:"login.v1"`
}{
	ModulesV1: tfeapi.ModuleV1Prefix,
	MotdV1:    "/api/terraform/motd",
	StateV2:   tfeapi.APIPrefixV2,
	TfeV2:     tfeapi.APIPrefixV2,
	TfeV21:    tfeapi.APIPrefixV2,
	TfeV22:    tfeapi.APIPrefixV2,
	LoginV1:   loginserver.Discovery,
})

type Service struct{}

func (Service) AddHandlers(r *mux.Router) {
	r.HandleFunc("/.well-known/terraform.json", func(w gohttp.ResponseWriter, r *gohttp.Request) {
		w.Header().Set("Content-type", "application/json")
		w.Write(discoveryPayload)
	})
}
