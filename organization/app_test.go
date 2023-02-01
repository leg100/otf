package organization

import (
	"strconv"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrganizationList(t *testing.T) {
	// create a dozen orgs
	var orgs []*Organization
	for i := 0; i < 12; i++ {
		org, err := NewOrganization(OrganizationCreateOptions{
			Name: otf.String(strconv.Itoa(i)),
		})
		require.NoError(t, err)
		orgs = append(orgs, org)
	}

	tests := []struct {
		name string
		opts otf.ListOptions
		// wanted total - should always be 12
		wantTotal int
		// wanted number of items on page
		wantItems int
	}{
		{
			name:      "default",
			opts:      otf.ListOptions{},
			wantTotal: 12,
			wantItems: otf.DefaultPageSize,
		},
		{
			name:      "second page",
			opts:      otf.ListOptions{PageNumber: 2},
			wantTotal: 12,
			wantItems: 12 - otf.DefaultPageSize,
		},
		{
			name:      "out of bounds",
			opts:      otf.ListOptions{PageNumber: 999, PageSize: 999},
			wantTotal: 12,
			wantItems: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := OrganizationListOptions{ListOptions: tt.opts}
			list := newOrganizationList(opts, orgs)
			assert.Equal(t, tt.wantTotal, list.TotalCount())
			assert.Equal(t, tt.wantItems, len(list.Items))
		})
	}
}
