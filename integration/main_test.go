package integration

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// The otfd daemon spawned in an integration test uses a self-signed cert.
	// The following environment variable instructs any Go program spawned in a
	// test, e.g. the terraform CLI, the otf agent, etc, to trust the
	// self-signed cert.
	// * Assign the *absolute* path to the SSL cert because Go program's working
	// directory may differ from the integration test directory.
	wd, err := os.Getwd()
	panicIfError(err)
	unset := setenv("SSL_CERT_FILE", filepath.Join(wd, "./fixtures/cert.pem"))
	defer unset()

	// Create dedicated home directory for duration of integration tests.
	// Terraform CLI and the `otf` CLI create various directories and dot files
	// in the home directory and we do not want to pollute the user's home
	// directory.
	homeDir, err := os.MkdirTemp("", "")
	panicIfError(err)
	defer func() {
		os.RemoveAll(homeDir)
	}()
	unset = setenv("HOME", homeDir)
	defer unset()

	// If HTTPS_PROXY has been defined then add it to the authoritative list of
	// environment variables so that processes, particularly terraform, spawed
	// in tests use the proxy. This can be very useful for caching repeated
	// downloads of terraform providers during tests.
	if proxy, ok := os.LookupEnv("HTTPS_PROXY"); ok {
		envs = append(envs, "HTTPS_PROXY="+proxy)
	}

	// Instruct terraform CLI to skip checks for new versions.
	unset = setenv("CHECKPOINT_DISABLE", "true")
	defer unset()

	// Ensure ~/.terraform.d exists - 'terraform login' has a bug whereby it tries to
	// persist the API token it receives to a temporary file in ~/.terraform.d but
	// fails if ~/.terraform.d doesn't exist yet. This only happens when
	// CHECKPOINT_DISABLE is set, because the checkpoint would otherwise handle
	// creating that directory first.
	os.MkdirAll(path.Join(os.Getenv("HOME"), ".terraform.d"), 0o755)

	os.Exit(m.Run())
}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// setenv sets an environment variable and returns a func to unset the variable.
// The environment variable is added to a shared slice, envs, for individual
// tests to use.
func setenv(name, value string) func() {
	err := os.Setenv(name, value)
	if err != nil {
		panic(err.Error())
	}
	envs = append(envs, name+"="+value)
	return func() {
		os.Unsetenv(name)
	}
}
