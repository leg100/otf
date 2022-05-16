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

func UnmarshalOrganizationListFromDB(pgresult interface{}) (organizations []*Organization, err error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var rows []OrganizationDBRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		org, err := unmarshalOrganizationDBRow(row)
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, org)
	}

	return organizations, nil
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

	return unmarshalOrganizationDBRow(row)
}

func unmarshalOrganizationDBRow(row OrganizationDBRow) (*Organization, error) {
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
