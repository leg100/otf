package otf

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	jsonapi "github.com/leg100/otf/http/dto"
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

// ConfigurationVersion is a representation of an uploaded or ingressed
// Terraform configuration in  A workspace must have at least one configuration
// version before any runs may be queued on it.
type ConfigurationVersion struct {
	id               string
	createdAt        time.Time
	autoQueueRuns    bool
	source           ConfigurationSource
	speculative      bool
	status           ConfigurationStatus
	statusTimestamps []ConfigurationVersionStatusTimestamp
	// Configuration Version belongs to a Workspace
	Workspace *Workspace
}

func (cv *ConfigurationVersion) ID() string                  { return cv.id }
func (cv *ConfigurationVersion) CreatedAt() time.Time        { return cv.createdAt }
func (cv *ConfigurationVersion) String() string              { return cv.id }
func (cv *ConfigurationVersion) AutoQueueRuns() bool         { return cv.autoQueueRuns }
func (cv *ConfigurationVersion) Source() ConfigurationSource { return cv.source }
func (cv *ConfigurationVersion) Speculative() bool           { return cv.speculative }
func (cv *ConfigurationVersion) Status() ConfigurationStatus { return cv.status }
func (cv *ConfigurationVersion) StatusTimestamps() []ConfigurationVersionStatusTimestamp {
	return cv.statusTimestamps
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

// ToJSONAPI assembles a JSONAPI DTO.
func (cv *ConfigurationVersion) ToJSONAPI(req *http.Request) any {
	obj := &jsonapi.ConfigurationVersion{
		ID:               cv.ID(),
		AutoQueueRuns:    cv.AutoQueueRuns(),
		Speculative:      cv.Speculative(),
		Source:           string(cv.Source()),
		Status:           string(cv.Status()),
		StatusTimestamps: &jsonapi.CVStatusTimestamps{},
		UploadURL:        fmt.Sprintf("/configuration-versions/%s/upload", cv.ID()),
	}
	for _, ts := range cv.StatusTimestamps() {
		switch ts.Status {
		case ConfigurationPending:
			obj.StatusTimestamps.QueuedAt = &ts.Timestamp
		case ConfigurationErrored:
			obj.StatusTimestamps.FinishedAt = &ts.Timestamp
		case ConfigurationUploaded:
			obj.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return obj
}

// ToJSONAPI assembles a JSONAPI DTO
func (l *ConfigurationVersionList) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.ConfigurationVersionList{
		Pagination: (*jsonapi.Pagination)(l.Pagination),
	}
	for _, item := range l.Items {
		dto.Items = append(dto.Items, item.ToJSONAPI(req).(*jsonapi.ConfigurationVersion))
	}
	return dto
}

// ConfigurationStatus represents a configuration version status.
type ConfigurationStatus string

// ConfigurationVersionList represents a list of configuration versions.
type ConfigurationVersionList struct {
	*Pagination
	Items []*ConfigurationVersion
}

// ConfigurationSource represents a source of a configuration version.
type ConfigurationSource string

type ConfigurationVersionStatusTimestamp struct {
	Status    ConfigurationStatus
	Timestamp time.Time
}

// ConfigurationVersionCreateOptions represents the options for creating a
// configuration version. See dto.ConfigurationVersionCreateOptions for more
// details.
type ConfigurationVersionCreateOptions struct {
	AutoQueueRuns *bool
	Speculative   *bool
}

type ConfigurationVersionService interface {
	Create(ctx context.Context, workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
	Get(ctx context.Context, id string) (*ConfigurationVersion, error)
	GetLatest(ctx context.Context, workspaceID string) (*ConfigurationVersion, error)
	List(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)

	// Upload handles verification and upload of the config tarball, updating
	// the config version upon success or failure.
	Upload(ctx context.Context, id string, config []byte) error

	// Download retrieves the config tarball for the given config version ID.
	Download(ctx context.Context, id string) ([]byte, error)
}

type ConfigurationVersionStore interface {
	// Creates a config version.
	CreateConfigurationVersion(ctx context.Context, cv *ConfigurationVersion) error
	// Get retrieves a config version.
	GetConfigurationVersion(ctx context.Context, opts ConfigurationVersionGetOptions) (*ConfigurationVersion, error)
	// GetConfig retrieves the config tarball for the given config version ID.
	GetConfig(ctx context.Context, id string) ([]byte, error)
	// List lists config versions for the given workspace.
	ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)
	// Delete deletes the config version from the store
	DeleteConfigurationVersion(ctx context.Context, id string) error
	// Upload uploads a config tarball for the given config version ID
	UploadConfigurationVersion(ctx context.Context, id string, fn func(cv *ConfigurationVersion, uploader ConfigUploader) error) error
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

	ListOptions
}

// NewConfigurationVersion creates a ConfigurationVersion object from scratch
func NewConfigurationVersion(workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		id:            NewID("cv"),
		createdAt:     CurrentTimestamp(),
		autoQueueRuns: DefaultAutoQueueRuns,
		status:        ConfigurationPending,
		source:        DefaultConfigurationSource,
	}

	if opts.AutoQueueRuns != nil {
		cv.autoQueueRuns = *opts.AutoQueueRuns
	}

	if opts.Speculative != nil {
		cv.speculative = *opts.Speculative
	}

	cv.Workspace = &Workspace{id: workspaceID}

	return &cv, nil
}
