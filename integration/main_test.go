package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// The otfd daemon spawned in an integration test uses a self-signed cert.
	// The following environment variable instructs any Go program spawned in a
	// test, e.g. the terraform CLI, the otf agent, etc, to trust the
	// self-signed cert.
	// Assign the *absolute* path to the SSL cert because Go program's working
	// directory may differ from the integration test directory.
	wd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	err = os.Setenv("SSL_CERT_FILE", filepath.Join(wd, "./fixtures/cert.pem"))
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		os.Unsetenv("SSL_CERT_FILE")
	}()
	os.Exit(m.Run())
}
