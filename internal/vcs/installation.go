package vcs

import (
	"errors"

	"github.com/google/go-github/v65/github"
)

var ErrAppNotFound = errors.New("github app not found")

type Installation struct {
	ID           int64
	AppID        int64
	Username     *string
	Organization *string
	HTMLURL      string
}

func (i Installation) String() string {
	if i.Organization != nil {
		return "org/" + *i.Organization
	}
	return "user/" + *i.Username
}

func NewInstallation(ghInstall *github.Installation) (Installation, error) {
	id := ghInstall.ID
	if id == nil {
		return Installation{}, errors.New("installation is missing an ID")
	}
	appID := ghInstall.AppID
	if appID == nil {
		return Installation{}, errors.New("installation is missing an app ID")
	}
	install := Installation{
		ID:      *id,
		AppID:   *appID,
		HTMLURL: ghInstall.GetHTMLURL(),
	}

	switch ghInstall.GetAccount().GetType() {
	case "Organization":
		install.Organization = ghInstall.GetAccount().Login
	case "User":
		install.Username = ghInstall.GetAccount().Login
	default:
		return Installation{}, errors.New("installation must have an account type of either username or organization")
	}

	return install, nil
}
