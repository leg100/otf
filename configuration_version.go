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
	ListConfigurationVersions(opts tfe.ConfigurationVersionListOptions) (*tfe.ConfigurationVersionList, error)
	UploadConfigurationVersion(id string, payload []byte) error
}

func NewConfigurationVersionID() string {
	return fmt.Sprintf("cv-%s", GenerateRandomString(16))
}
