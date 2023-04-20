package integration

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func createBrowserCtx(t *testing.T) context.Context {
	t.Helper()

	headless := true
	if v, ok := os.LookupEnv("OTF_E2E_HEADLESS"); ok {
		var err error
		headless, err = strconv.ParseBool(v)
		if err != nil {
			panic("cannot parse OTF_E2E_HEADLESS")
		}
	}

	// must create an allocator before creating the browser
	allocator, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", headless),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)
	t.Cleanup(cancel)

	// now create the browser
	ctx, cancel := chromedp.NewContext(allocator)
	t.Cleanup(cancel)

	// Ensure ~/.terraform.d exists - 'terraform login' has a bug whereby it tries to
	// persist the API token it receives to a temporary file in ~/.terraform.d but
	// fails if ~/.terraform.d doesn't exist yet. This only happens when
	// CHECKPOINT_DISABLE is set, because the checkpoint would otherwise handle
	// creating that directory first...
	os.MkdirAll(path.Join(os.Getenv("HOME"), ".terraform.d"), 0o755)

	return ctx
}

func workspaceURL(hostname, org, name string) string {
	return "https://" + hostname + "/app/organizations/" + org + "/workspaces/" + name
}

func organizationURL(hostname, org string) string {
	return "https://" + hostname + "/app/organizations/" + org
}

// newRootModule creates a terraform root module, returning its directory path
func newRootModule(t *testing.T, hostname, organization, workspace string, additionalConfig ...string) string {
	t.Helper()

	config := fmt.Sprintf(`
terraform {
  backend "remote" {
	hostname = "%s"
	organization = "%s"

	workspaces {
	  name = "%s"
	}
  }
}
resource "null_resource" "e2e" {}
`, hostname, organization, workspace)
	for _, cfg := range additionalConfig {
		config += "\n"
		config += cfg
	}

	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "main.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	return root
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	return contents
}
