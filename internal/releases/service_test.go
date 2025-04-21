package releases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_latestCheck(t *testing.T) {
	tests := []struct {
		name          string
		currentLatest string
		newLatest     string
		lastCheck     time.Time
		wantCheck     bool
	}{
		{
			"perform check and expect newer version",
			"1.2.3",
			"1.2.4",
			time.Now().Add(-time.Hour * 48),
			true,
		},
		{
			"perform check and expect same version",
			"1.2.3",
			"1.2.3",
			time.Now().Add(-time.Hour * 48),
			true,
		},
		{
			"skip check because previous check performed an hour ago",
			"",
			"",
			time.Now().Add(-time.Hour),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				db: &testDB{currentLatest: tt.currentLatest, lastCheck: tt.lastCheck},
			}
			before, after, err := s.check(context.Background(), &testEngine{latestVersion: tt.newLatest})
			assert.NoError(t, err)
			assert.Equal(t, tt.currentLatest, before)
			assert.Equal(t, tt.newLatest, after)
		})
	}
}
