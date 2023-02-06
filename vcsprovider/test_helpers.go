package vcsprovider

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/stretchr/testify/require"
)

func NewTestVCSProvider(t *testing.T, org *otf.Organization) *VCSProvider {
	factory := &factory{inmem.NewTestCloudService()}
	provider, err := factory.new(createOptions{
		Organization: org.Name(),
		// unit tests require a legitimate cloud name to avoid invalid foreign
		// key error upon insert/update
		Cloud: "github",
		Name:  uuid.NewString(),
		Token: uuid.NewString(),
	})
	require.NoError(t, err)
	return provider
}
