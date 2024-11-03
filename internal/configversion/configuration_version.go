// Package configversion handles terraform configurations.
package configversion

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

const (
	DefaultAutoQueueRuns = true

	// List all available configuration version statuses.
	ConfigurationErrored  ConfigurationStatus = "errored"
	ConfigurationPending  ConfigurationStatus = "pending"
	ConfigurationUploaded ConfigurationStatus = "uploaded"

	// Default maximum config size is 10mb.
	DefaultConfigMaxSize int64 = 1024 * 1024 * 10

	ConfigVersionKind     = "cv"
	IngressAttributesKind = "ia"
)

type (
	// ConfigurationVersion is a representation of an uploaded or ingressed
	// Terraform configuration.
	ConfigurationVersion struct {
		resource.ID

		CreatedAt         time.Time
		AutoQueueRuns     bool
		Source            Source
		Speculative       bool
		Status            ConfigurationStatus
		StatusTimestamps  []ConfigurationVersionStatusTimestamp
		WorkspaceID       resource.ID
		IngressAttributes *IngressAttributes
	}

	// CreateOptions represents the options for creating a
	// configuration version. See jsonapi.CreateOptions for more
	// details.
	CreateOptions struct {
		AutoQueueRuns *bool
		Speculative   *bool
		Source        Source
		*IngressAttributes
	}

	// ConfigurationStatus represents a configuration version status.
	ConfigurationStatus string

	ConfigurationVersionStatusTimestamp struct {
		Status    ConfigurationStatus
		Timestamp time.Time
	}

	// ConfigUploader uploads a config
	ConfigUploader interface {
		// Upload uploads the config tarball and returns a status indicating success
		// or failure.
		Upload(ctx context.Context, config []byte) (ConfigurationStatus, error)
		// SetErrored sets the config version status to 'errored' in the store.
		SetErrored(ctx context.Context) error
	}

	// ConfigurationVersionGetOptions are options for retrieving a single config
	// version. Either ID *or* WorkspaceID must be specfiied.
	ConfigurationVersionGetOptions struct {
		// ID of config version to retrieve
		ID *resource.ID

		// Get latest config version for this workspace ID
		WorkspaceID *resource.ID

		// A list of relations to include
		Include *string `schema:"include"`
	}

	// ListOptions are options for paginating and filtering a
	// list of configuration versions
	ListOptions struct {
		// A list of relations to include
		Include *string `schema:"include"`

		resource.PageOptions
	}

	IngressAttributes struct {
		// ID     string
		Branch string
		// CloneURL          string
		// CommitMessage     string
		CommitSHA string
		CommitURL string
		// CompareURL        string
		Repo              string
		IsPullRequest     bool
		OnDefaultBranch   bool
		PullRequestNumber int
		PullRequestURL    string
		PullRequestTitle  string
		// PullRequestBody   string
		Tag             string
		SenderUsername  string
		SenderAvatarURL string
		SenderHTMLURL   string
	}
)

// NewConfigurationVersion creates a ConfigurationVersion object from scratch
func NewConfigurationVersion(workspaceID resource.ID, opts CreateOptions) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		ID:            resource.NewID(ConfigVersionKind),
		CreatedAt:     internal.CurrentTimestamp(nil),
		AutoQueueRuns: DefaultAutoQueueRuns,
		Source:        DefaultSource,
		WorkspaceID:   workspaceID,
	}
	cv.updateStatus(ConfigurationPending)

	if opts.Source != "" {
		cv.Source = opts.Source
	}
	if opts.AutoQueueRuns != nil {
		cv.AutoQueueRuns = *opts.AutoQueueRuns
	}
	if opts.Speculative != nil {
		cv.Speculative = *opts.Speculative
	}
	if opts.IngressAttributes != nil {
		cv.IngressAttributes = opts.IngressAttributes
	}
	return &cv, nil
}

func (cv *ConfigurationVersion) StatusTimestamp(status ConfigurationStatus) (time.Time, error) {
	for _, sts := range cv.StatusTimestamps {
		if sts.Status == status {
			return sts.Timestamp, nil
		}
	}
	return time.Time{}, internal.ErrStatusTimestampNotFound
}

func (cv *ConfigurationVersion) AddStatusTimestamp(status ConfigurationStatus, timestamp time.Time) {
	cv.StatusTimestamps = append(cv.StatusTimestamps, ConfigurationVersionStatusTimestamp{
		Status:    status,
		Timestamp: timestamp,
	})
}

// Upload saves the config to the db and updates status accordingly.
func (cv *ConfigurationVersion) Upload(ctx context.Context, config []byte, uploader ConfigUploader) error {
	// upload config and set status depending on success
	status, err := uploader.Upload(ctx, config)
	if err != nil {
		return err
	}
	cv.Status = status

	return nil
}

func (cv *ConfigurationVersion) updateStatus(status ConfigurationStatus) {
	cv.Status = status
	cv.StatusTimestamps = append(cv.StatusTimestamps, ConfigurationVersionStatusTimestamp{
		Status:    status,
		Timestamp: internal.CurrentTimestamp(nil),
	})
}
