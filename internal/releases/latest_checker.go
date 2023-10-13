package releases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const latestEndpoint = "https://api.releases.hashicorp.com/v1/releases/terraform/latest"

// latestChecker checks for a new latest release of terraform.
type latestChecker struct {
	endpoint string
}

func (c latestChecker) check(last time.Time) (string, error) {
	// skip check if already checked within last 24 hours
	if last.After(time.Now().Add(-24 * time.Hour)) {
		return "", nil
	}
	// check releases endpoint
	resp, err := http.Get(c.endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%s return non-200 status code: %s", c.endpoint, resp.Status)
	}
	// decode endpoint response
	var release struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.Version, nil
}
