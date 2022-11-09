package otf

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
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

func NewTestVCSProvider(t *testing.T, organization *Organization, cloud Cloud) *VCSProvider {
	provider, err := NewVCSProvider(cloud, VCSProviderCreateOptions{
		Name:             uuid.NewString(),
		Token:            uuid.NewString(),
		OrganizationName: organization.Name(),
	})
	require.NoError(t, err)
	provider.cloud = cloud
	return provider
}

func NewTestVCSRepo(provider *VCSProvider) *VCSRepo {
	identifier := uuid.NewString()
	return &VCSRepo{
		Identifier: identifier,
		HTTPURL:    "http://fake-cloud.org/" + identifier,
		ProviderID: provider.ID(),
		Branch:     "master",
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

// NewTestTarball creates a tarball (.tar) consisting of files respectively populated with the
// given contents. The files are assigned random names with the terraform file
// extension appended (.tf)
func NewTestTarball(t *testing.T, contents ...string) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

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

	tw.Close()
	return buf.Bytes()
}

// NewTestTarGZ wraps NewTestTarball, creating a .tar.gz instead.
func NewTestTarGZ(t *testing.T, contents ...string) []byte {
	return compress(t, NewTestTarball(t, contents...))
}

// compress runs gzip compression on the input bytes
func compress(t *testing.T, b []byte) []byte {
	var buf bytes.Buffer
	compressor := gzip.NewWriter(&buf)
	_, err := io.Copy(compressor, bytes.NewReader(b))
	require.NoError(t, err)
	compressor.Close()
	return buf.Bytes()
}
