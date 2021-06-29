package ots

import (
	"fmt"

	tfe "github.com/leg100/go-tfe"
)

const (
	DefaultAutoQueueRuns = true
)

type ConfigurationVersionService interface {
	CreateConfigurationVersion(opts *tfe.ConfigurationVersionCreateOptions) (*tfe.ConfigurationVersion, error)
	GetConfigurationVersion(id string) (*tfe.ConfigurationVersion, error)
	ListConfigurationVersions(opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)
	UploadConfigurationVersion(id string, payload []byte) error
}

type ConfigurationVersionList struct {
	*Pagination
	Items []*tfe.ConfigurationVersion
}

// ConfigurationVersionListOptions represents the options for listing organizations.
type ConfigurationVersionListOptions struct {
	ListOptions
}

func NewConfigurationVersionID() string {
	return fmt.Sprintf("cv-%s", GenerateRandomString(16))
}
