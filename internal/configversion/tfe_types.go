package configversion

import (
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi/types"
)

// TFEConfigurationVersion is an uploaded or ingressed Terraform configuration. A workspace
// must have at least one configuration version before any runs may be queued on it.
type TFEConfigurationVersion struct {
	ID               resource.TfeID         `jsonapi:"primary,configuration-versions"`
	AutoQueueRuns    bool                   `jsonapi:"attribute" json:"auto-queue-runs"`
	Error            string                 `jsonapi:"attribute" json:"error"`
	ErrorMessage     string                 `jsonapi:"attribute" json:"error-message"`
	Source           string                 `jsonapi:"attribute" json:"source"`
	Speculative      bool                   `jsonapi:"attribute" json:"speculative"`
	Status           string                 `jsonapi:"attribute" json:"status"`
	StatusTimestamps *TFECVStatusTimestamps `jsonapi:"attribute" json:"status-timestamps"`
	UploadURL        string                 `jsonapi:"attribute" json:"upload-url"`

	// Relations
	IngressAttributes *TFEIngressAttributes `jsonapi:"relationship" json:"ingress-attributes"`
}

// CVStatusTimestamps holds the timestamps for individual configuration version
// statuses.
type TFECVStatusTimestamps struct {
	FinishedAt *time.Time `json:"finished-at,omitempty"`
	QueuedAt   *time.Time `json:"queued-at,omitempty"`
	StartedAt  *time.Time `json:"started-at,omitempty"`
}

// ConfigurationVersionList represents a list of configuration versions.
type TFEConfigurationVersionList struct {
	*types.Pagination
	Items []*TFEConfigurationVersion
}

// ConfigurationVersionCreateOptions represents the options for creating a
// configuration version.
type TFEConfigurationVersionCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,configuration-versions"`

	// When true, runs are queued automatically when the configuration version
	// is uploaded.
	AutoQueueRuns *bool `jsonapi:"attribute" json:"auto-queue-runs,omitempty"`

	// When true, this configuration version can only be used for planning.
	Speculative *bool `jsonapi:"attribute" json:"speculative,omitempty"`
}

type TFEIngressAttributes struct {
	ID                resource.TfeID `jsonapi:"primary,ingress-attributes"`
	Branch            string         `jsonapi:"attribute" json:"branch"`
	CloneURL          string         `jsonapi:"attribute" json:"clone-url"`
	CommitMessage     string         `jsonapi:"attribute" json:"commit-message"`
	CommitSHA         string         `jsonapi:"attribute" json:"commit-sha"`
	CommitURL         string         `jsonapi:"attribute" json:"commit-url"`
	CompareURL        string         `jsonapi:"attribute" json:"compare-url"`
	Identifier        string         `jsonapi:"attribute" json:"identifier"`
	IsPullRequest     bool           `jsonapi:"attribute" json:"is-pull-request"`
	OnDefaultBranch   bool           `jsonapi:"attribute" json:"on-default-branch"`
	PullRequestNumber int            `jsonapi:"attribute" json:"pull-request-number"`
	PullRequestURL    string         `jsonapi:"attribute" json:"pull-request-url"`
	PullRequestTitle  string         `jsonapi:"attribute" json:"pull-request-title"`
	PullRequestBody   string         `jsonapi:"attribute" json:"pull-request-body"`
	Tag               string         `jsonapi:"attribute" json:"tag"`
	SenderUsername    string         `jsonapi:"attribute" json:"sender-username"`
	SenderAvatarURL   string         `jsonapi:"attribute" json:"sender-avatar-url"`
	SenderHTMLURL     string         `jsonapi:"attribute" json:"sender-html-url"`
}
