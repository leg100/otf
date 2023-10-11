package releases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf/internal/semver"
)

const latestEndpoint = "https://api.releases.hashicorp.com/v1/releases/terraform/latest"

// latestChecker checks for a new latest release of terraform.
type latestChecker struct {
	endpoint string
}

// check takes the last time a version was checked and the current latest
// version we have on record, and returns whether it checked or not, and if
// there is a new latest release then "newer" is populated with its version.
func (c latestChecker) check(last time.Time, v string) (newer string, checked bool, err error) {
	// skip check if already checked within last 24 hours
	if last.After(time.Now().Add(-24 * time.Hour)) {
		return "", false, nil
	}
	// check releases endpoint
	resp, err := http.Get(c.endpoint)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", false, fmt.Errorf("%s return non-200 status code: %s", c.endpoint, resp.Status)
	}
	// decode endpoint response
	var release struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", false, err
	}
	// check if newer
	if n := semver.Compare(release.Version, v); n > 0 {
		// is newer
		return release.Version, true, nil
	}
	// not newer
	return "", true, nil
}
