package ots

import (
	"fmt"

	"github.com/google/jsonapi"
	tfe "github.com/hashicorp/go-tfe"
)

const (
	DefaultAutoQueueRuns = true
)

// ConfigurationVersion is a representation of an uploaded or ingressed
// Terraform configuration in TFE. A workspace must have at least one
// configuration version before any runs may be queued on it.
type ConfigurationVersion struct {
	ID               string                  `jsonapi:"primary,configuration-versions"`
	AutoQueueRuns    bool                    `jsonapi:"attr,auto-queue-runs"`
	Error            string                  `jsonapi:"attr,error"`
	ErrorMessage     string                  `jsonapi:"attr,error-message"`
	Source           tfe.ConfigurationSource `jsonapi:"attr,source"`
	Speculative      bool                    `jsonapi:"attr,speculative "`
	Status           tfe.ConfigurationStatus `jsonapi:"attr,status"`
	StatusTimestamps *tfe.CVStatusTimestamps `jsonapi:"attr,status-timestamps"`
	UploadURL        string                  `jsonapi:"attr,upload-url"`
}

func (cv *ConfigurationVersion) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("/api/v2/configuration-versions/%s", cv.ID),
	}
}

type ConfigurationVersionService interface {
	CreateConfigurationVersion(opts *tfe.ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
	GetConfigurationVersion(id string) (*ConfigurationVersion, error)
	ListConfigurationVersions(opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)
	UploadConfigurationVersion(id string, payload []byte) error
}

func NewConfigurationVersionID() string {
	return fmt.Sprintf("cv-%s", GenerateRandomString(16))
}
