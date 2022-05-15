package otf

import (
	"encoding/json"
	"time"
)

type OrganizationDBRow struct {
	OrganizationID  string    `json:"organization_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Name            string    `json:"name"`
	SessionRemember int       `json:"session_remember"`
	SessionTimeout  int       `json:"session_timeout"`
	FullCount       int       `json:"full_count"`
}

func UnmarshalOrganizationFromDB(pgresult interface{}) (*Organization, error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var row OrganizationDBRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}

	org := Organization{
		ID: row.OrganizationID,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Name:            row.Name,
		SessionRemember: row.SessionRemember,
		SessionTimeout:  row.SessionTimeout,
	}

	return &org, nil
}
