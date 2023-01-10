package otf

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
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

type NewTestWorkspaceOption func(*WorkspaceCreateOptions)

func AutoApply() NewTestWorkspaceOption {
	return func(opts *WorkspaceCreateOptions) {
		opts.AutoApply = Bool(true)
	}
}

func WithRepo(repo *WorkspaceRepo) NewTestWorkspaceOption {
	return func(opts *WorkspaceCreateOptions) {
		opts.Repo = repo
	}
}

func NewTestWorkspace(t *testing.T, org *Organization, opts ...NewTestWorkspaceOption) *Workspace {
	var createOpts WorkspaceCreateOptions
	for _, o := range opts {
		o(&createOpts)
	}
	if createOpts.Name == "" {
		createOpts.Name = uuid.NewString()
	}
	ws, err := NewWorkspace(org, createOpts)
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

func NewTestOwners(t *testing.T, org *Organization, opts ...NewTeamOption) *Team {
	return NewTeam("owners", org, opts...)
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
		cloudConfig: CloudConfig{
			Name:     "fake-cloud",
			Hostname: "fake-cloud.org",
		},
	}
}

func NewTestWorkspaceRepo(provider *VCSProvider, hook *Webhook) *WorkspaceRepo {
	return &WorkspaceRepo{
		ProviderID: provider.ID(),
		Branch:     "master",
		Identifier: hook.Identifier,
	}
}

func NewTestRepo() *Repo {
	identifier := uuid.NewString() + "/" + uuid.NewString()
	return &Repo{
		Identifier: identifier,
		Branch:     "master",
	}
}

func NewTestModuleRepo(provider, name string) *Repo {
	identifier := fmt.Sprintf("%s/terraform-%s-%s", uuid.New(), provider, name)
	return &Repo{
		Identifier: identifier,
		Branch:     "master",
	}
}

func NewTestWebhook(repo *Repo, cloudConfig CloudConfig) *Webhook {
	return &Webhook{
		WebhookID:   uuid.New(),
		VCSID:       "123",
		Secret:      "secret",
		Identifier:  repo.Identifier,
		cloudConfig: cloudConfig,
	}
}

func NewTestModule(org *Organization, opts ...NewTestModuleOption) *Module {
	createOpts := CreateModuleOptions{
		Organization: org,
		Provider:     uuid.NewString(),
		Name:         uuid.NewString(),
	}
	mod := NewModule(createOpts)
	for _, o := range opts {
		o(mod)
	}
	return mod
}

type NewTestModuleOption func(*Module)

func WithModuleStatus(status ModuleStatus) NewTestModuleOption {
	return func(mod *Module) {
		mod.status = status
	}
}

func WithModuleVersion(version string, status ModuleVersionStatus) NewTestModuleOption {
	return func(mod *Module) {
		mod.Add(NewTestModuleVersion(mod, version, status))
	}
}

func WithModuleRepo() NewTestModuleOption {
	return func(mod *Module) {
		mod.repo = &ModuleRepo{}
	}
}

func NewTestModuleVersion(mod *Module, version string, status ModuleVersionStatus) *ModuleVersion {
	createOpts := CreateModuleVersionOptions{
		ModuleID: mod.ID(),
		Version:  version,
	}
	modver := NewModuleVersion(createOpts)
	modver.status = status
	return modver
}

func NewTestCloudConfig(cloud Cloud) CloudConfig {
	return CloudConfig{
		Name:     "fake-cloud",
		Hostname: "fake-cloud.org",
		Cloud:    cloud,
	}
}

func NewTestAgentToken(t *testing.T, org *Organization) *AgentToken {
	token, err := NewAgentToken(CreateAgentTokenOptions{
		OrganizationName: org.Name(),
		Description:      "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
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
