package ots

import "net/url"

var _ Paginated = (*StateVersionList)(nil)

type StateVersionList struct {
	Items []*StateVersion
	StateVersionListOptions
}

// StateVersionListOptions represents the options for listing state versions.
type StateVersionListOptions struct {
	ListOptions
	StateVersionListFilters
}

// StateVersionListFilters filters state version list by org and workspace
type StateVersionListFilters struct {
	Organization *string `schema:"filter[organization][name],required"`
	Workspace    *string `schema:"filter[workspace][name],required"`
}

func (l *StateVersionList) GetItems() interface{} {
	return l.Items
}

func (l *StateVersionList) GetListOptions() ListOptions {
	return l.ListOptions
}

func (l *StateVersionList) GetPath() string {
	query := url.Values{}
	if err := encoder.Encode(l.StateVersionListFilters, query); err != nil {
		panic(err.Error())
	}

	return (&url.URL{Path: "/api/v2/state-versions", RawQuery: query.Encode()}).String()
}
