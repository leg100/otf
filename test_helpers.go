package otf

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

func NewTestWorkspaceRepo(provider VCSProvider) *WorkspaceRepo {
	return &WorkspaceRepo{
		ProviderID: provider.ID(),
		Branch:     "master",
		Identifier: "leg100/" + uuid.NewString(),
	}
}

func NewTestCloudConfig(c cloud.Cloud) cloud.Config {
	return cloud.Config{
		Name:     "fake-cloud",
		Hostname: "fake-cloud.org",
		Cloud:    c,
	}
}

// NewTestTarball creates a tarball (.tar.gz) consisting of files respectively populated with the
// given contents. The files are assigned random names with the terraform file
// extension appended (.tf)
func NewTestTarball(t *testing.T, contents ...string) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for _, body := range contents {
		header := &tar.Header{
			Name: uuid.NewString() + ".tf",
			Mode: 0o600,
			Size: int64(len(body)),
		}
		err := tw.WriteHeader(header)
		require.NoError(t, err)

		_, err = tw.Write([]byte(body))
		require.NoError(t, err)
	}

	err := tw.Close()
	require.NoError(t, err)
	err = gw.Close()
	require.NoError(t, err)

	return buf.Bytes()
}
