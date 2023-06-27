package integration

import (
	"testing"

	"github.com/leg100/otf/internal/repo"
)

// TestRepoDB tests repo package's interaction with a (real) database.
func TestRepoDB(t *testing.T) {
	integrationTest(t)

	// Call out to integration test situated within 'repo' package, which then
	// calls private methods
	repo.TestDB(t)
}
