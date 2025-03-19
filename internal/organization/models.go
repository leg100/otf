// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package organization

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

type Model struct {
	OrganizationID             resource.ID
	CreatedAt                  pgtype.Timestamptz
	UpdatedAt                  pgtype.Timestamptz
	Name                       Name
	SessionRemember            pgtype.Int4
	SessionTimeout             pgtype.Int4
	Email                      pgtype.Text
	CollaboratorAuthPolicy     pgtype.Text
	AllowForceDeleteWorkspaces pgtype.Bool
	CostEstimationEnabled      pgtype.Bool
}

type TokenModel struct {
	OrganizationTokenID resource.ID
	CreatedAt           pgtype.Timestamptz
	OrganizationName    pgtype.Text
	Expiry              pgtype.Timestamptz
}
