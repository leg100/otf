package otf

import (
	"context"
	"time"
)

type ConfigurationVersionService interface {
	CreateConfigurationVersion(ctx context.Context, workspaceID string, opts ConfigurationVersionCreateOptions) (ConfigurationVersion, error)
	// CloneConfigurationVersion creates a new configuration version using the
	// config tarball of an existing configuration version.
	CloneConfigurationVersion(ctx context.Context, cvID string, opts ConfigurationVersionCreateOptions) (ConfigurationVersion, error)
	GetConfigurationVersion(ctx context.Context, id string) (ConfigurationVersion, error)
	GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (ConfigurationVersion, error)

	// Upload handles verification and upload of the config tarball, updating
	// the config version upon success or failure.
	UploadConfig(ctx context.Context, id string, config []byte) error

	// Download retrieves the config tarball for the given config version ID.
	DownloadConfig(ctx context.Context, id string) ([]byte, error)
}

type ConfigurationVersion interface {
	ID() string
	CreatedAt() time.Time
	String() string
	AutoQueueRuns() bool
	Speculative() bool

	IngressAttributes() *IngressAttributes
}

// ConfigurationVersionCreateOptions represents the options for creating a
// configuration version. See jsonapi.ConfigurationVersionCreateOptions for more
// details.
type ConfigurationVersionCreateOptions struct {
	AutoQueueRuns *bool
	Speculative   *bool
	*IngressAttributes
}

type IngressAttributes struct {
	// ID     string
	Branch string
	// CloneURL          string
	// CommitMessage     string
	CommitSHA string
	// CommitURL         string
	// CompareURL        string
	Identifier      string
	IsPullRequest   bool
	OnDefaultBranch bool
	// PullRequestNumber int
	// PullRequestURL    string
	// PullRequestTitle  string
	// PullRequestBody   string
	// Tag               string
	// SenderUsername    string
	// SenderAvatarURL   string
	// SenderHTMLURL     string
}
