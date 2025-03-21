package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgo(t *testing.T) {
	tests := []struct {
		name string
		ago  time.Duration
		want string
	}{
		{"now", 0, "0s ago"},
		{"5 seconds ago", 5 * time.Second, "5s ago"},
		{"5 minutes ago", 5 * time.Minute, "5m ago"},
		{"5 hours ago", 5 * time.Hour, "5h ago"},
		{"5 days ago", 5 * 24 * time.Hour, "5d ago"},
		{"5 weeks ago", 5 * 7 * 24 * time.Hour, "35d ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			got := Ago(now, now.Add(-tt.ago))
			assert.Equal(t, tt.want, got)
		})
	}
}
