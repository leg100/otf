package dynamiccreds

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type Handlers struct {
	hostnameService *internal.HostnameService
	publicKey       jwk.Key
}

type wellKnownConfig struct {
	// IssuerURL is the identity of the provider, and the string it uses to sign
	// ID tokens with. For example "https://accounts.google.com". This value MUST
	// match ID tokens exactly.
	IssuerURL string `json:"issuer"`
	// JWKSURL is the endpoint used by the provider to advertise public keys to
	// verify issued ID tokens. This endpoint is polled as new keys are made
	// available.
	JWKSURL string `json:"jwks_uri"`
	// Algorithms, if provided, indicate a list of JWT algorithms allowed to sign
	// ID tokens. If not provided, this defaults to the algorithms advertised by
	// the JWK endpoint, then the set of algorithms supported by this package.
	Algorithms []string `json:"id_token_signing_alg_values_supported"`
	// REQUIRED. JSON array containing a list of the OAuth 2.0 response_type values that this OP supports. Dynamic OpenID Providers MUST support the code, id_token, and the id_token token Response Type values.
	ResponseTypes []string `json:"response_types_supported"`
	// OPTIONAL. JSON array containing a list of the Claim Types that the OpenID Provider supports. These Claim Types are described in Section 5.6 of OpenID Connect Core 1.0 [OpenID.Core]. Values defined by this specification are normal, aggregated, and distributed. If omitted, the implementation supports only normal Claims.
	Claims []string `json:"claims_supported"`
	// RECOMMENDED. JSON array containing a list of the OAuth 2.0 [RFC6749] scope values that this server supports. The server MUST support the openid scope value. Servers MAY choose not to advertise some supported scope values even when this parameter is used, although those defined in [OpenID.Core] SHOULD be listed, if supported.
	Scopes []string `json:"scopes_supported"`
	// REQUIRED. JSON array containing a list of the Subject Identifier types that this OP supports. Valid types include pairwise and public.
	SubjectTypes []string `json:"subject_types_supported"`
}

func (h *Handlers) addHandlers(r *mux.Router) {
	r.HandleFunc("/.well-known/openid-configuration", h.wellKnown).Methods("GET")
	r.HandleFunc("/.well-known/jwks", h.jwks).Methods("GET")
}

func (h *Handlers) wellKnown(w http.ResponseWriter, r *http.Request) {
	cfg := wellKnownConfig{
		IssuerURL:     h.hostnameService.URL(""),
		JWKSURL:       h.hostnameService.URL("/.well-known/jwks"),
		Algorithms:    []string{"RS256"},
		Scopes:        []string{"openid"},
		SubjectTypes:  []string{"public"},
		ResponseTypes: []string{"id_token"},
		Claims: []string{
			"sub",
			"aud",
			"exp",
			"iat",
			"nbf",
			"iss",
			"jti",
			"terraform_organization_name",
			"terraform_workspace_name",
			"terraform_workspace_id",
			"terraform_full_workspace",
			"terraform_run_id",
			"terraform_run_phase",
		},
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) jwks(w http.ResponseWriter, r *http.Request) {
	set := jwk.NewSet()
	set.AddKey(h.publicKey)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(set); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
