package sql

import "github.com/leg100/otf"

func convertSessionComposite(row Sessions) *otf.Session {
	return &otf.Session{
		Token: *row.Token,
		Timestamps: otf.Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Expiry: row.Expiry,
		UserID: *row.UserID,
		SessionData: otf.SessionData{
			Address: *row.Address,
		},
	}
}
