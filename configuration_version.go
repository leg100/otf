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
	ID string `db:"configuration_version_id" jsonapi:"primary,configuration-versions"`

	Timestamps

	AutoQueueRuns    bool
	Source           ConfigurationSource
	Speculative      bool
	Status           ConfigurationStatus
	StatusTimestamps []ConfigurationVersionStatusTimestamp

	// Configuration Version belongs to a Workspace
	Workspace *Workspace `db:"workspaces"`
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
	Upload(id string, payload []byte) error
	Download(id string) ([]byte, error)
}

type ConfigurationVersionStore interface {
	Create(run *ConfigurationVersion) (*ConfigurationVersion, error)
	Get(opts ConfigurationVersionGetOptions) (*ConfigurationVersion, error)
	GetConfig(ctx context.Context, id string) ([]byte, error)
	List(workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)
	Update(id string, fn func(*ConfigurationVersion, ConfigurationVersionUpdater) error) error
	Delete(id string) error
}

// ConfigurationVersionUpdater updates a config version within a transaction
type ConfigurationVersionUpdater interface {
	UpdateStatus(ctx context.Context, status ConfigurationStatus) (ConfigurationVersionStatusTimestamp, error)
	SaveConfig(ctx context.Context, config []byte) error
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

// ConfigurationVersionFactory creates ConfigurationVersion objects
type ConfigurationVersionFactory struct {
	WorkspaceService WorkspaceService
}

func (cv *ConfigurationVersion) GetID() string  { return cv.ID }
func (cv *ConfigurationVersion) String() string { return cv.ID }

// Upload saves the config to the db and updates status accordingly.
func (cv *ConfigurationVersion) Upload(ctx context.Context, config []byte, updater ConfigurationVersionUpdater) error {
	if cv.Status != ConfigurationPending {
		return fmt.Errorf("attempted to upload configuration version with non-pending status: %s", cv.Status)
	}

	if err := updater.SaveConfig(ctx, config); err != nil {
		return err
	}

	if _, err := updater.UpdateStatus(ctx, ConfigurationUploaded); err != nil {
		return err
	}

	return nil
}

// NewConfigurationVersion creates a ConfigurationVersion object from scratch
func (f *ConfigurationVersionFactory) NewConfigurationVersion(workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		ID:            NewID("cv"),
		Timestamps:    NewTimestamps(),
		AutoQueueRuns: DefaultAutoQueueRuns,
		Status:        ConfigurationPending,
		Source:        DefaultConfigurationSource,
	}

	if opts.AutoQueueRuns != nil {
		cv.AutoQueueRuns = *opts.AutoQueueRuns
	}

	if opts.Speculative != nil {
		cv.Speculative = *opts.Speculative
	}

	ws, err := f.WorkspaceService.Get(context.Background(), WorkspaceSpec{ID: &workspaceID})
	if err != nil {
		return nil, err
	}
	cv.Workspace = ws

	return &cv, nil
}
