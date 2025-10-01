package dynamiccreds

import (
	"testing"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_assignKidHeader(t *testing.T) {
	svc, err := NewService(Options{
		PublicKeyPath:  "./testdata/public_key.pem",
		PrivateKeyPath: "./testdata/private_key.pem",
	})
	require.NoError(t, err)

	// Check that a `kid` header has been assigned to both private and public
	// keys.
	assert.NotEmpty(t, svc.handlers.publicKey.KeyID())
	assert.NotEmpty(t, svc.privateKey.KeyID())
}

func TestService_GenerateToken(t *testing.T) {
	svc, err := NewService(Options{
		PublicKeyPath:  "./testdata/public_key.pem",
		PrivateKeyPath: "./testdata/private_key.pem",
	})
	require.NoError(t, err)

	got, err := svc.GenerateToken(
		"https://issuer.url",
		organization.NewTestName(t),
		resource.NewTfeID(resource.WorkspaceKind),
		"dev",
		resource.NewTfeID(resource.RunKind),
		run.PlanPhase,
		"aud",
	)
	require.NoError(t, err)

	msg, err := jws.Parse(got)
	require.NoError(t, err)

	// check `kid` header has been added to jwt.
	assert.Equal(t, 1, len(msg.Signatures()))
	protected := msg.Signatures()[0].ProtectedHeaders()
	assert.NotEmpty(t, protected.KeyID())
}
