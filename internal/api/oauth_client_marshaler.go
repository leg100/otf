package api

import (
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/vcsprovider"
)

func (m *jsonapiMarshaler) toOAuthClient(from *vcsprovider.VCSProvider) *types.OAuthClient {
	to := &types.OAuthClient{
		ID:        from.ID,
		CreatedAt: from.CreatedAt,
		// Only github via github.com is supported currently, so hardcode these values.
		ServiceProvider: types.ServiceProviderGithub,
		APIURL:          githubAPIURL,
		HTTPURL:         githubHTTPURL,
		// OTF has no corresponding concept of an OAuthToken, so just use the
		// VCS provider ID (the go-tfe integration tests we use expect
		// at least an ID).
		OAuthTokens: []*types.OAuthToken{
			{ID: from.ID},
		},
		Organization: &types.Organization{Name: from.Organization},
	}
	// A name is mandatory in OTF, but in go-tfe integration tests it is
	// optional; therefore OTF takes an empty string to mean nil (this is
	// necessary in order to pass the go-tfe integration tests).
	if from.Name != "" {
		to.Name = &from.Name
	}
	return to
}
