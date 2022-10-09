package html

import "github.com/leg100/otf"

type UserList struct {
	Items []*otf.User
	opts  otf.UserListOptions
}

func (l UserList) OrganizationName() string {
	if l.opts.OrganizationName == nil {
		return ""
	}
	return *l.opts.OrganizationName
}

func (l UserList) TeamName() string {
	if l.opts.TeamName == nil {
		return ""
	}
	return *l.opts.TeamName
}
