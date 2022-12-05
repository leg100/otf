package otf

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestOrganization(t *testing.T) *Organization {
	org, err := NewOrganization(OrganizationCreateOptions{
		Name: String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

func NewTestWorkspace(t *testing.T, org *Organization, opts WorkspaceCreateOptions) *Workspace {
	if opts.Name == "" {
		opts.Name = uuid.NewString()
	}
	ws, err := NewWorkspace(org, opts)
	require.NoError(t, err)
	return ws
}

func NewTestConfigurationVersion(t *testing.T, ws *Workspace, opts ConfigurationVersionCreateOptions) *ConfigurationVersion {
	cv, err := NewConfigurationVersion(ws.ID(), opts)
	require.NoError(t, err)
	return cv
}

func NewTestUser(t *testing.T, opts ...NewUserOption) *User {
	return NewUser(uuid.NewString(), opts...)
}

func NewTestTeam(t *testing.T, org *Organization, opts ...NewTeamOption) *Team {
	return NewTeam(uuid.NewString(), org, opts...)
}

func NewTestSession(t *testing.T, userID string, opts ...NewSessionOption) *Session {
	session, err := NewSession(userID, "127.0.0.1")
	require.NoError(t, err)

	for _, o := range opts {
		o(session)
	}

	return session
}

func NewTestVCSProvider(t *testing.T, organization *Organization) *VCSProvider {
	return &VCSProvider{
		id:               NewID("vcs"),
		createdAt:        CurrentTimestamp(),
		name:             uuid.NewString(),
		token:            uuid.NewString(),
		organizationName: organization.Name(),
		cloudConfig:      GithubDefaults(),
	}
}

func NewTestWorkspaceRepo(provider *VCSProvider) *WorkspaceRepo {
	return &WorkspaceRepo{
		ProviderID: provider.ID(),
		Branch:     "master",
		Webhook: NewTestWebhook(NewTestRepo(), CloudConfig{
			Name:     "fake-cloud",
			Hostname: "fake-cloud.org",
		}),
	}
}

func NewTestRepo() *Repo {
	identifier := uuid.NewString() + "/" + uuid.NewString()
	return &Repo{
		Identifier: identifier,
		HTTPURL:    "http://fake-cloud.org/" + identifier,
		Branch:     "master",
	}
}

func NewTestWebhook(repo *Repo, cloudConfig CloudConfig) *Webhook {
	return &Webhook{
		WebhookID:   uuid.New(),
		VCSID:       "123",
		Secret:      "secret",
		Identifier:  repo.Identifier,
		HTTPURL:     repo.HTTPURL,
		cloudConfig: cloudConfig,
	}
}

// NewTestTarball creates a tarball (.tar.gz) consisting of files respectively populated with the
// given contents. The files are assigned random names with the terraform file
// extension appended (.tf)
func NewTestTarball(t *testing.T, contents ...string) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for _, body := range contents {
		header := &tar.Header{
			Name: uuid.NewString() + ".tf",
			Mode: 0o600,
			Size: int64(len(body)),
		}
		err := tw.WriteHeader(header)
		require.NoError(t, err)

		_, err = tw.Write([]byte(body))
		require.NoError(t, err)
	}

	err := tw.Close()
	require.NoError(t, err)
	err = gw.Close()
	require.NoError(t, err)

	return buf.Bytes()
}
