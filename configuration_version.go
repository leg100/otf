package otf

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	DefaultAutoQueueRuns       = true
	DefaultConfigurationSource = "tfe-api"

	//List all available configuration version statuses.
	ConfigurationErrored  ConfigurationStatus = "errored"
	ConfigurationPending  ConfigurationStatus = "pending"
	ConfigurationUploaded ConfigurationStatus = "uploaded"
)

var (
	ErrInvalidConfigurationVersionGetOptions = errors.New("invalid configuration version get options")
)

// ConfigurationStatus represents a configuration version status.
type ConfigurationStatus string

// ConfigurationVersionList represents a list of configuration versions.
type ConfigurationVersionList struct {
	*Pagination
	Items []*ConfigurationVersion
}

// ConfigurationSource represents a source of a configuration version.
type ConfigurationSource string

// ConfigurationVersion is a representation of an uploaded or ingressed
// Terraform configuration in  A workspace must have at least one
// configuration version before any runs may be queued on it.
type ConfigurationVersion struct {
	ID string `jsonapi:"primary,configuration-versions"`

	Timestamps

	autoQueueRuns    bool
	source           ConfigurationSource
	speculative      bool
	status           ConfigurationStatus
	statusTimestamps []ConfigurationVersionStatusTimestamp

	// Configuration Version belongs to a Workspace
	Workspace *Workspace
}

type ConfigurationVersionStatusTimestamp struct {
	Status    ConfigurationStatus
	Timestamp time.Time
}

// ConfigurationVersionCreateOptions represents the options for creating a
// configuration version.
type ConfigurationVersionCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,configuration-versions"`

	// When true, runs are queued automatically when the configuration version
	// is uploaded.
	AutoQueueRuns *bool `jsonapi:"attr,auto-queue-runs,omitempty"`

	// When true, this configuration version can only be used for planning.
	Speculative *bool `jsonapi:"attr,speculative,omitempty"`
}

type ConfigurationVersionService interface {
	Create(workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
	Get(id string) (*ConfigurationVersion, error)
	GetLatest(workspaceID string) (*ConfigurationVersion, error)
	List(workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)

	// Upload handles verification and upload of the config tarball, updating
	// the config version upon success or failure.
	Upload(id string, config []byte) error

	// Download retrieves the config tarball for the given config version ID.
	Download(id string) ([]byte, error)
}

type ConfigurationVersionStore interface {
	// Creates a config version.
	Create(run *ConfigurationVersion) (*ConfigurationVersion, error)

	// Get retrieves a config version.
	Get(opts ConfigurationVersionGetOptions) (*ConfigurationVersion, error)

	// GetConfig retrieves the config tarball for the given config version ID.
	GetConfig(ctx context.Context, id string) ([]byte, error)

	// List lists config versions for the given workspace.
	List(workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)

	// Delete deletes the config version from the store
	Delete(id string) error

	// Upload uploads a config tarball for the given config version ID
	Upload(ctx context.Context, id string, fn func(cv *ConfigurationVersion, uploader ConfigUploader) error) error
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
}

// ConfigurationVersionListOptions are options for paginating and filtering a
// list of configuration versions
type ConfigurationVersionListOptions struct {
	// A list of relations to include
	Include *string `schema:"include"`

	ListOptions
}

func (cv *ConfigurationVersion) AutoQueueRuns() bool         { return cv.autoQueueRuns }
func (cv *ConfigurationVersion) GetID() string               { return cv.ID }
func (cv *ConfigurationVersion) Source() ConfigurationSource { return cv.source }
func (cv *ConfigurationVersion) Speculative() bool           { return cv.speculative }
func (cv *ConfigurationVersion) Status() ConfigurationStatus { return cv.status }
func (cv *ConfigurationVersion) String() string              { return cv.ID }
func (cv *ConfigurationVersion) StatusTimestamps() []ConfigurationVersionStatusTimestamp {
	return cv.statusTimestamps
}

func (cv *ConfigurationVersion) AddStatusTimestamp(status ConfigurationStatus, timestamp time.Time) {
	cv.statusTimestamps = append(cv.statusTimestamps, ConfigurationVersionStatusTimestamp{
		Status:    status,
		Timestamp: timestamp,
	})
}

func (cv *ConfigurationVersion) ShallowNest() {
	cv.statusTimestamps = nil
	cv.Workspace = &Workspace{ID: cv.Workspace.ID}
}

// Upload saves the config to the db and updates status accordingly.
func (cv *ConfigurationVersion) Upload(ctx context.Context, config []byte, uploader ConfigUploader) error {
	if cv.status != ConfigurationPending {
		return fmt.Errorf("attempted to upload configuration version with non-pending status: %s", cv.status)
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
