package github

import "github.com/google/go-github/v65/github"

type Installation struct {
	*github.Installation
}

func (i *Installation) String() string {
	if i.GetAccount().GetType() == "Organization" {
		return "org/" + i.GetAccount().GetLogin()
	}
	return "user/" + i.GetAccount().GetLogin()
}
