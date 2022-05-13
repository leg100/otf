package sql

import "github.com/leg100/otf"

func convertTokenComposite(row Tokens) *otf.Token {
	return &otf.Token{
		ID: *row.UserID,
		Timestamps: otf.Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Token:       *row.TokenID,
		Description: *row.Description,
		UserID:      *row.UserID,
	}
}
