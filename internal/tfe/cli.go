package tfe

import "net/http"

const (
	// headerSource is an http header providing the source of the API call
	headerSource = "X-Terraform-Integration"
	// headerSourceCLI is an http header value for headerSource that indicates
	// the source of the API call is the terraform CLI
	headerSourceCLI = "cloud"
)

func IsTerraformCLI(r *http.Request) bool {
	return r.Header.Get(headerSource) == headerSourceCLI
}
