package otf

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

func NewTestOrganization(t *testing.T) *Organization {
	org, err := NewOrganization(OrganizationCreateOptions{
		Name: String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

type NewTestWorkspaceOption func(*CreateWorkspaceOptions)

func AutoApply() NewTestWorkspaceOption {
	return func(opts *CreateWorkspaceOptions) {
		opts.AutoApply = Bool(true)
	}
}

func WithRepo(repo *WorkspaceRepo) NewTestWorkspaceOption {
	return func(opts *CreateWorkspaceOptions) {
		opts.Repo = repo
	}
}

func NewTestWorkspace(t *testing.T, org *Organization, opts ...NewTestWorkspaceOption) *Workspace {
	createOpts := CreateWorkspaceOptions{
		Name:         String(uuid.NewString()),
		Organization: String(org.Name()),
	}
	for _, o := range opts {
		o(&createOpts)
	}
	ws, err := NewWorkspace(createOpts)
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

func NewTestRegistrySession(t *testing.T, org *Organization, opts ...NewTestRegistrySessionOption) *RegistrySession {
	session, err := NewRegistrySession(org.Name())
	require.NoError(t, err)

	for _, o := range opts {
		o(session)
	}

	return session
}

type NewTestRegistrySessionOption func(*RegistrySession)

func OverrideTestRegistrySessionExpiry(expiry time.Time) NewTestRegistrySessionOption {
	return func(session *RegistrySession) {
		session.expiry = expiry
	}
}

func NewTestVCSProvider(t *testing.T, organization *Organization) *VCSProvider {
	return &VCSProvider{
		id:           NewID("vcs"),
		createdAt:    CurrentTimestamp(),
		name:         uuid.NewString(),
		token:        uuid.NewString(),
		organization: organization.Name(),
		cloudConfig: cloud.Config{
			Name:     "fake-cloud",
			Hostname: "fake-cloud.org",
		},
	}
}

func NewTestWorkspaceRepo(provider *VCSProvider) *WorkspaceRepo {
	return &WorkspaceRepo{
		ProviderID: provider.ID(),
		Branch:     "master",
		Identifier: "leg100/" + uuid.NewString(),
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

func NewTestCloudConfig(c cloud.Cloud) cloud.Config {
	return cloud.Config{
		Name:     "fake-cloud",
		Hostname: "fake-cloud.org",
		Cloud:    c,
	}
}

func NewTestAgentToken(t *testing.T, org *Organization) *AgentToken {
	token, err := NewAgentToken(CreateAgentTokenOptions{
		Organization: org.Name(),
		Description:  "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
}

func NewTestVariable(t *testing.T, ws *Workspace, opts CreateVariableOptions) *Variable {
	v, err := NewVariable(ws.ID(), opts)
	require.NoError(t, err)
	return v
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
