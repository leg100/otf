package ots

type StateVersionList struct {
	*Pagination
	Items []*StateVersion
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
