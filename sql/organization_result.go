package sql

import (
	"time"

	"github.com/leg100/otf"
)

type organizationListResult interface {
	GetOrganizations() []Organizations

	GetFullCount() *int
}

type organizationRow struct {
	OrganizationID  string    `json:"organization_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Name            string    `json:"name"`
	SessionRemember *int32    `json:"session_remember"`
	SessionTimeout  *int32    `json:"session_timeout"`
}

func convertOrganizationComposite(row Organizations) *otf.Organization {
	organization := otf.Organization{
		ID:              *row.GetOrganizationID(),
		Name:            *row.GetName(),
		Timestamps:      convertTimestamps(row),
		SessionRemember: int(*row.GetSessionRemember()),
		SessionTimeout:  int(*row.GetSessionTimeout()),
	}

	return &organization
}
