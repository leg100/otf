package html

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestCurrentOrganization(t *testing.T) {
	tests := []struct {
		name    string
		content any
		want    *string
	}{
		{
			name:    "string field of anonymous struct",
			content: struct{ Organization string }{Organization: "acme-corp"},
			want:    otf.String("acme-corp"),
		},
		{
			name:    "string field of anonymous struct inside interface",
			content: any(struct{ Organization string }{Organization: "acme-corp"}),
			want:    otf.String("acme-corp"),
		},
		{
			name:    "currentOrganization implementation",
			content: &testOrganizationResource{org: "acme-corp"},
			want:    otf.String("acme-corp"),
		},
		{
			name:    "currentOrganization implementation inside interface",
			content: any(&testOrganizationResource{org: "acme-corp"}),
			want:    otf.String("acme-corp"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CurrentOrganization(tt.content))
		})
	}
}

type testOrganizationResource struct {
	org string
}

func (t *testOrganizationResource) OrganizationName() string { return t.org }
