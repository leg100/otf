// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package github

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type GithubApp struct {
	GithubAppID   pgtype.Int8
	WebhookSecret pgtype.Text
	PrivateKey    pgtype.Text
	Slug          pgtype.Text
	Organization  pgtype.Text
}
