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

func (i *Installation) Organization() *string {
	if i.GetAccount().GetType() == "Organization" {
		org := i.GetAccount().GetLogin()
		return &org
	}
	return nil
}

func (i *Installation) Username() *string {
	if i.GetAccount().GetType() == "User" {
		username := i.GetAccount().GetLogin()
		return &username
	}
	return nil
}
