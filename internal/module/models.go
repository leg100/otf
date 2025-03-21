// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package module

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

type ModuleVersionModel struct {
	ModuleVersionID resource.TfeID
	Version         pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	Status          pgtype.Text
	StatusError     pgtype.Text
	ModuleID        resource.TfeID
}
