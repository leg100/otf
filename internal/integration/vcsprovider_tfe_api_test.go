package integration

import (
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_VCSProviderTFEAPI tests those parts of the TFE API for VCS
// Providers (what TFE calls OAuth Clients) that aren't covered by the go-tfe
// integration tests.
func TestIntegration_VCSProviderTFEAPI(t *testing.T) {
	integrationTest(t)

	t.Run("create gitlab provider", func(t *testing.T) {
		daemon, org, ctx := setup(t)

		_, token := daemon.createToken(t, ctx, nil)

		client, err := tfe.NewClient(&tfe.Config{
			Address: daemon.System.URL("/"),
			Token:   string(token),
		})
		require.NoError(t, err)

		created, err := client.OAuthClients.Create(ctx, org.Name.String(), tfe.OAuthClientCreateOptions{
			OAuthToken:      internal.Ptr("my-pat"),
			HTTPURL:         internal.Ptr("http://gitlab.com"),
			APIURL:          internal.Ptr("http://gitlab.com/api/v4"),
			ServiceProvider: internal.Ptr(tfe.ServiceProviderGitlab),
		})
		require.NoError(t, err)

		// Check that what was created can be retrieved and that the attributes
		// match
		retrieved, err := client.OAuthClients.Read(ctx, created.ID)
		require.NoError(t, err)

		assert.Equal(t, retrieved.ID, created.ID)
		assert.Equal(t, tfe.ServiceProviderGitlab, retrieved.ServiceProvider)
		assert.Equal(t, "http://gitlab.com", retrieved.HTTPURL)
		assert.Equal(t, "http://gitlab.com/api/v4", retrieved.APIURL)
	})
}
