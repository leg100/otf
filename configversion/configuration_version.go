package configversion

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf"
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

// ConfigurationVersion is a representation of an uploaded or ingressed
// Terraform configuration in  A workspace must have at least one configuration
// version before any runs may be queued on it.
type ConfigurationVersion struct {
	id                string
	createdAt         time.Time
	autoQueueRuns     bool
	source            ConfigurationSource
	speculative       bool
	status            ConfigurationStatus
	statusTimestamps  []ConfigurationVersionStatusTimestamp
	workspaceID       string
	ingressAttributes *otf.IngressAttributes
}

// NewConfigurationVersion creates a ConfigurationVersion object from scratch
func NewConfigurationVersion(workspaceID string, opts otf.ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		id:            otf.NewID("cv"),
		createdAt:     otf.CurrentTimestamp(),
		autoQueueRuns: DefaultAutoQueueRuns,
		source:        otf.DefaultConfigurationSource,
		workspaceID:   workspaceID,
	}
	cv.updateStatus(ConfigurationPending)

	if opts.AutoQueueRuns != nil {
		cv.autoQueueRuns = *opts.AutoQueueRuns
	}
	if opts.Speculative != nil {
		cv.speculative = *opts.Speculative
	}
	if opts.IngressAttributes != nil {
		cv.ingressAttributes = opts.IngressAttributes
	}
	return &cv, nil
}

func (cv *ConfigurationVersion) ID() string                  { return cv.id }
func (cv *ConfigurationVersion) CreatedAt() time.Time        { return cv.createdAt }
func (cv *ConfigurationVersion) String() string              { return cv.id }
func (cv *ConfigurationVersion) AutoQueueRuns() bool         { return cv.autoQueueRuns }
func (cv *ConfigurationVersion) Source() ConfigurationSource { return cv.source }
func (cv *ConfigurationVersion) Speculative() bool           { return cv.speculative }
func (cv *ConfigurationVersion) Status() ConfigurationStatus { return cv.status }
func (cv *ConfigurationVersion) WorkspaceID() string         { return cv.workspaceID }
func (cv *ConfigurationVersion) Key() string                 { return cacheKey(cv.id) }
func (cv *ConfigurationVersion) StatusTimestamps() []ConfigurationVersionStatusTimestamp {
	return cv.statusTimestamps
}

func (cv *ConfigurationVersion) IngressAttributes() *otf.IngressAttributes {
	return cv.ingressAttributes
}

func (cv *ConfigurationVersion) StatusTimestamp(status ConfigurationStatus) (time.Time, error) {
	for _, sts := range cv.statusTimestamps {
		if sts.Status == status {
			return sts.Timestamp, nil
		}
	}
	return time.Time{}, otf.ErrStatusTimestampNotFound
}

func (cv *ConfigurationVersion) AddStatusTimestamp(status ConfigurationStatus, timestamp time.Time) {
	cv.statusTimestamps = append(cv.statusTimestamps, ConfigurationVersionStatusTimestamp{
		Status:    status,
		Timestamp: timestamp,
	})
}

// Upload saves the config to the db and updates status accordingly.
func (cv *ConfigurationVersion) Upload(ctx context.Context, config []byte, uploader ConfigUploader) error {
	if cv.status != ConfigurationPending {
		return fmt.Errorf("cannot upload config for a configuration version with non-pending status: %s", cv.status)
	}

	// check config untars successfully and set errored status if not

	// upload config and set status depending on success
	status, err := uploader.Upload(ctx, config)
	if err != nil {
		return err
	}
	cv.status = status

	return nil
}

func (cv *ConfigurationVersion) updateStatus(status ConfigurationStatus) {
	cv.status = status
	cv.statusTimestamps = append(cv.statusTimestamps, ConfigurationVersionStatusTimestamp{
		Status:    status,
		Timestamp: otf.CurrentTimestamp(),
	})
}

// ConfigurationStatus represents a configuration version status.
type ConfigurationStatus string

// ConfigurationVersionList represents a list of configuration versions.
type ConfigurationVersionList struct {
	*otf.Pagination
	Items []*ConfigurationVersion
}

// ConfigurationSource represents a source of a configuration version.
type ConfigurationSource string

type ConfigurationVersionStatusTimestamp struct {
	Status    ConfigurationStatus
	Timestamp time.Time
}

// ConfigUploader uploads a config
type ConfigUploader interface {
	// Upload uploads the config tarball and returns a status indicating success
	// or failure.
	Upload(ctx context.Context, config []byte) (ConfigurationStatus, error)
	// SetErrored sets the config version status to 'errored' in the store.
	SetErrored(ctx context.Context) error
}

// ConfigurationVersionGetOptions are options for retrieving a single config
// version. Either ID *or* WorkspaceID must be specfiied.
type ConfigurationVersionGetOptions struct {
	// ID of config version to retrieve
	ID *string

	// Get latest config version for this workspace ID
	WorkspaceID *string

	// A list of relations to include
	Include *string `schema:"include"`
}

// ConfigurationVersionListOptions are options for paginating and filtering a
// list of configuration versions
type ConfigurationVersionListOptions struct {
	// A list of relations to include
	Include *string `schema:"include"`

	otf.ListOptions
}