package configversion

import (
	"time"

	"github.com/leg100/otf"
)

// ConfigurationVersion is a representation of an uploaded or ingressed
// Terraform configuration in TFE. A workspace must have at least one
// configuration version before any runs may be queued on it.
type jsonapiConfigurationVersion struct {
	ID               string                     `jsonapi:"primary,configuration-versions"`
	AutoQueueRuns    bool                       `jsonapi:"attr,auto-queue-runs"`
	Error            string                     `jsonapi:"attr,error"`
	ErrorMessage     string                     `jsonapi:"attr,error-message"`
	Source           string                     `jsonapi:"attr,source"`
	Speculative      bool                       `jsonapi:"attr,speculative "`
	Status           string                     `jsonapi:"attr,status"`
	StatusTimestamps *jsonapiCVStatusTimestamps `jsonapi:"attr,status-timestamps"`
	UploadURL        string                     `jsonapi:"attr,upload-url"`
}

// CVStatusTimestamps holds the timestamps for individual configuration version
// statuses.
type jsonapiCVStatusTimestamps struct {
	FinishedAt *time.Time `json:"finished-at,omitempty"`
	QueuedAt   *time.Time `json:"queued-at,omitempty"`
	StartedAt  *time.Time `json:"started-at,omitempty"`
}

// ConfigurationVersionList represents a list of configuration versions.
type jsonapiConfigurationVersionList struct {
	*otf.Pagination
	Items []*ConfigurationVersion
}

// ConfigurationVersionCreateOptions represents the options for creating a
// configuration version.
type jsonapiConfigurationVersionCreateOptions struct {
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

type jsonapiIngressAttributes struct {
	ID                string `jsonapi:"primary,ingress-attributes"`
	Branch            string `jsonapi:"attr,branch"`
	CloneURL          string `jsonapi:"attr,clone-url"`
	CommitMessage     string `jsonapi:"attr,commit-message"`
	CommitSHA         string `jsonapi:"attr,commit-sha"`
	CommitURL         string `jsonapi:"attr,commit-url"`
	CompareURL        string `jsonapi:"attr,compare-url"`
	Identifier        string `jsonapi:"attr,identifier"`
	IsPullRequest     bool   `jsonapi:"attr,is-pull-request"`
	OnDefaultBranch   bool   `jsonapi:"attr,on-default-branch"`
	PullRequestNumber int    `jsonapi:"attr,pull-request-number"`
	PullRequestURL    string `jsonapi:"attr,pull-request-url"`
	PullRequestTitle  string `jsonapi:"attr,pull-request-title"`
	PullRequestBody   string `jsonapi:"attr,pull-request-body"`
	Tag               string `jsonapi:"attr,tag"`
	SenderUsername    string `jsonapi:"attr,sender-username"`
	SenderAvatarURL   string `jsonapi:"attr,sender-avatar-url"`
	SenderHTMLURL     string `jsonapi:"attr,sender-html-url"`

	// Links
	Links map[string]interface{} `jsonapi:"links,omitempty"`
}
