package ots

import (
	"fmt"
	"net/url"
)

var DefaultArchiveHost = "localhost:8080"

func GetPlanLogsUrl(id string) string {
	return (&url.URL{
		Scheme: "https",
		Host:   DefaultArchiveHost,
		Path:   fmt.Sprintf("/plans/%s/logs", id),
	}).String()
}
