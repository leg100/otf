package ots

import (
	"fmt"

	tfe "github.com/leg100/go-tfe"
)

const (
	DefaultAutoQueueRuns       = true
	DefaultConfigurationSource = "tfe-api"
)

type ConfigurationVersionService interface {
	CreateConfigurationVersion(workspaceID string, opts *tfe.ConfigurationVersionCreateOptions) (*tfe.ConfigurationVersion, error)
	GetConfigurationVersion(id string) (*tfe.ConfigurationVersion, error)
	ListConfigurationVersions(workspaceID string, opts tfe.ConfigurationVersionListOptions) (*tfe.ConfigurationVersionList, error)
	UploadConfigurationVersion(id string, payload []byte) error
}

func NewConfigurationVersionID() string {
	return fmt.Sprintf("cv-%s", GenerateRandomString(16))
}
