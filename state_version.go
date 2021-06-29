package ots

import (
	"fmt"

	"github.com/leg100/go-tfe"
)

type StateVersionService interface {
	CreateStateVersion(workspaceID string, opts *tfe.StateVersionCreateOptions) (*tfe.StateVersion, error)
	ListStateVersions(orgName, workspaceName string, opts StateVersionListOptions) (*StateVersionList, error)
	CurrentStateVersion(workspaceID string) (*tfe.StateVersion, error)
	GetStateVersion(id string) (*tfe.StateVersion, error)
	DownloadStateVersion(id string) ([]byte, error)
}

func NewStateVersionID() string {
	return fmt.Sprintf("sv-%s", GenerateRandomString(16))
}
