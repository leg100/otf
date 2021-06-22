package ots

import (
	"fmt"
)

type ConfigurationVersionList struct {
	Items []*ConfigurationVersion
	ConfigurationVersionListOptions
	Workspace string
}

// ConfigurationVersionListOptions represents the options for listing organizations.
type ConfigurationVersionListOptions struct {
	ListOptions
}

var _ Paginated = (*ConfigurationVersionList)(nil)

func (l *ConfigurationVersionList) GetItems() interface{} {
	return l.Items
}

func (l *ConfigurationVersionList) GetListOptions() ListOptions {
	return l.ListOptions
}

func (l *ConfigurationVersionList) GetPath() string {
	return fmt.Sprintf("/api/v2/workspaces/%s/configuration-versions", l.Workspace)
}
