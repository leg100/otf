package ots

type ConfigurationVersionList struct {
	*Pagination
	Items []*ConfigurationVersion
}

// ConfigurationVersionListOptions represents the options for listing organizations.
type ConfigurationVersionListOptions struct {
	ListOptions
}
