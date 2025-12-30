// Package configversion handles terraform configurations.
package configversion

import (
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
)

const (
	DefaultAutoQueueRuns = true

	// List all available configuration version statuses.
	ConfigurationErrored  ConfigurationStatus = "errored"
	ConfigurationPending  ConfigurationStatus = "pending"
	ConfigurationUploaded ConfigurationStatus = "uploaded"

	// Default maximum config size is 10mb.
	DefaultConfigMaxSize int64 = 1024 * 1024 * 10
)

type (
	// ConfigurationVersion is a representation of an uploaded or ingressed
	// Terraform configuration.
	ConfigurationVersion struct {
		ID                resource.TfeID
		CreatedAt         time.Time
		AutoQueueRuns     bool
		Source            source.Source
		Speculative       bool
		Status            ConfigurationStatus
		StatusTimestamps  []StatusTimestamp
		WorkspaceID       resource.TfeID
		IngressAttributes *IngressAttributes
	}

	// CreateOptions represents the options for creating a
	// configuration version. See jsonapi.CreateOptions for more
	// details.
	CreateOptions struct {
		AutoQueueRuns *bool
		Speculative   *bool
		Source        source.Source
		// CreatedAt overrides the time the config was created at - for testing
		// purposes only.
		CreatedAt *time.Time

		*IngressAttributes
	}

	// ConfigurationStatus represents a configuration version status.
	ConfigurationStatus string

	StatusTimestamp struct {
		ConfigurationVersionID resource.TfeID `db:"-"`
		Status                 ConfigurationStatus
		Timestamp              time.Time
	}

	// ConfigurationVersionGetOptions are options for retrieving a single config
	// version. Either ID *or* WorkspaceID must be specfiied.
	ConfigurationVersionGetOptions struct {
		// ID of config version to retrieve
		ID *resource.TfeID

		// Get latest config version for this workspace ID
		WorkspaceID *resource.TfeID

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
		Branch                 string
		CommitSHA              string
		Repo                   vcs.Repo
		IsPullRequest          bool
		OnDefaultBranch        bool
		ConfigurationVersionID resource.TfeID
		CommitURL              string
		PullRequestNumber      int
		PullRequestURL         string
		PullRequestTitle       string
		Tag                    string
		SenderUsername         string
		SenderAvatarURL        string
		SenderHTMLURL          string
	}
)

// NewConfigurationVersion creates a ConfigurationVersion object from scratch
func NewConfigurationVersion(workspaceID resource.TfeID, opts CreateOptions) *ConfigurationVersion {
	cv := ConfigurationVersion{
		ID:            resource.NewTfeID(resource.ConfigVersionKind),
		CreatedAt:     internal.CurrentTimestamp(nil),
		AutoQueueRuns: DefaultAutoQueueRuns,
		Source:        source.Default,
		WorkspaceID:   workspaceID,
	}
	cv.updateStatus(ConfigurationPending)

	if opts.CreatedAt != nil {
		cv.CreatedAt = *opts.CreatedAt
	}
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
	return &cv
}

// GetID implements resource.deleteableResource
func (cv *ConfigurationVersion) GetID() resource.TfeID { return cv.ID }

func (cv *ConfigurationVersion) StatusTimestamp(status ConfigurationStatus) (time.Time, error) {
	for _, sts := range cv.StatusTimestamps {
		if sts.Status == status {
			return sts.Timestamp, nil
		}
	}
	return time.Time{}, internal.ErrStatusTimestampNotFound
}

func (cv *ConfigurationVersion) updateStatus(status ConfigurationStatus) {
	cv.Status = status
	cv.StatusTimestamps = append(cv.StatusTimestamps, StatusTimestamp{
		ConfigurationVersionID: cv.ID,
		Status:                 status,
		Timestamp:              internal.CurrentTimestamp(nil),
	})
}

// LogValue implements slog.LogValuer.
func (cv *ConfigurationVersion) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", cv.ID.String()),
		slog.Time("created", cv.CreatedAt),
	)
}
