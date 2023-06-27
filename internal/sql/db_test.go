package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetDefaultMaxConnections(t *testing.T) {
	tests := []struct {
		name    string
		connstr string
		want    string
	}{
		{"empty dsn", "", "pool_max_conns=20"},
		{"non-empty dsn", "user=louis host=localhost", "user=louis host=localhost pool_max_conns=20"},
		{"postgres url", "postgres:///otf", "postgres:///otf?pool_max_conns=20"},
		{"postgresql url", "postgresql:///otf", "postgresql:///otf?pool_max_conns=20"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setDefaultMaxConnections(tt.connstr, 20)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
