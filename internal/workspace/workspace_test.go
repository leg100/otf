package workspace

import (
	"errors"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkspace(t *testing.T) {
	tests := []struct {
		name string
		opts CreateOptions
		want error
	}{
		{
			name: "default",
			opts: CreateOptions{
				Name:         internal.String("my-workspace"),
				Organization: internal.String("my-org"),
			},
		},
		{
			name: "missing name",
			opts: CreateOptions{
				Organization: internal.String("my-org"),
			},
			want: internal.ErrRequiredName,
		},
		{
			name: "missing organization",
			opts: CreateOptions{
				Name: internal.String("my-workspace"),
			},
			want: internal.ErrRequiredOrg,
		},
		{
			name: "invalid name",
			opts: CreateOptions{
				Name: internal.String("%*&^"),
			},
			want: internal.ErrInvalidName,
		},
		{
			name: "bad terraform version",
			opts: CreateOptions{
				Name:             internal.String("my-workspace"),
				Organization:     internal.String("my-org"),
				TerraformVersion: internal.String("1,2,0"),
			},
			want: internal.ErrInvalidTerraformVersion,
		},
		{
			name: "unsupported terraform version",
			opts: CreateOptions{
				Name:             internal.String("my-workspace"),
				Organization:     internal.String("my-org"),
				TerraformVersion: internal.String("0.14.0"),
			},
			want: internal.ErrUnsupportedTerraformVersion,
		},
		{
			name: "specifying both tags regex and trigger patterns",
			opts: CreateOptions{
				Name:            internal.String("my-workspace"),
				Organization:    internal.String("my-org"),
				TriggerPatterns: []string{"/foo/**/*.tf"},
				ConnectOptions: &ConnectOptions{
					TagsRegex: internal.String("\\d+"),
				},
			},
			want: ErrTagsRegexAndTriggerPatterns,
		},
		{
			name: "specifying both tags regex and always trigger",
			opts: CreateOptions{
				Name:          internal.String("my-workspace"),
				Organization:  internal.String("my-org"),
				AlwaysTrigger: internal.Bool(true),
				ConnectOptions: &ConnectOptions{
					TagsRegex: internal.String("\\d+"),
				},
			},
			want: ErrTagsRegexAndAlwaysTrigger,
		},
		{
			name: "specifying both trigger patterns and always trigger",
			opts: CreateOptions{
				Name:            internal.String("my-workspace"),
				Organization:    internal.String("my-org"),
				AlwaysTrigger:   internal.Bool(true),
				TriggerPatterns: []string{"/foo/**/*.tf"},
			},
			want: ErrTriggerPatternsAndAlwaysTrigger,
		},
		{
			name: "invalid trigger pattern",
			opts: CreateOptions{
				Name:            internal.String("my-workspace"),
				Organization:    internal.String("my-org"),
				TriggerPatterns: []string{"/foo/[**/*.tf"},
			},
			want: ErrInvalidTriggerPattern,
		},
		{
			name: "invalid tags regex",
			opts: CreateOptions{
				Name:         internal.String("my-workspace"),
				Organization: internal.String("my-org"),
				ConnectOptions: &ConnectOptions{
					RepoPath:      internal.String("leg100/otf"),
					VCSProviderID: internal.String("vcs-123"),
					TagsRegex:     internal.String("{**"),
				},
			},
			want: ErrInvalidTagsRegex,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWorkspace(tt.opts)
			assert.True(t, errors.Is(err, tt.want), "got: %s", err)
		})
	}
}

func TestWorkspace_UpdateError(t *testing.T) {
	tests := []struct {
		name string
		ws   *Workspace
		opts UpdateOptions
		want error
	}{
		{
			name: "invalid name",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name: internal.String("%*&^"),
			},
			want: internal.ErrInvalidName,
		},
		{
			name: "bad terraform version",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name:             internal.String("my-workspace"),
				TerraformVersion: internal.String("1,2,0"),
			},
			want: internal.ErrInvalidTerraformVersion,
		},
		{
			name: "unsupported terraform version",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name:             internal.String("my-workspace"),
				TerraformVersion: internal.String("0.14.0"),
			},
			want: internal.ErrUnsupportedTerraformVersion,
		},
		{
			name: "specifying both tags regex and trigger patterns",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name:            internal.String("my-workspace"),
				TriggerPatterns: []string{"/foo/**/*.tf"},
				ConnectOptions: &ConnectOptions{
					TagsRegex: internal.String("\\d+"),
				},
			},
			want: ErrTagsRegexAndTriggerPatterns,
		},
		{
			name: "specifying both tags regex and always trigger",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name:          internal.String("my-workspace"),
				AlwaysTrigger: internal.Bool(true),
				ConnectOptions: &ConnectOptions{
					TagsRegex: internal.String("\\d+"),
				},
			},
			want: ErrTagsRegexAndAlwaysTrigger,
		},
		{
			name: "specifying both trigger patterns and always trigger",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name:            internal.String("my-workspace"),
				AlwaysTrigger:   internal.Bool(true),
				TriggerPatterns: []string{"/foo/**/*.tf"},
			},
			want: ErrTriggerPatternsAndAlwaysTrigger,
		},
		{
			name: "invalid trigger pattern",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name:            internal.String("my-workspace"),
				TriggerPatterns: []string{"/foo/[**/*.tf"},
			},
			want: ErrInvalidTriggerPattern,
		},
		{
			name: "invalid tags regex",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name: internal.String("my-workspace"),
				ConnectOptions: &ConnectOptions{
					RepoPath:      internal.String("leg100/otf"),
					VCSProviderID: internal.String("vcs-123"),
					TagsRegex:     internal.String("{**"),
				},
			},
			want: ErrInvalidTagsRegex,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.ws.Update(tt.opts)
			assert.True(t, errors.Is(err, tt.want), "got: %s", err)
		})
	}
}

func TestWorkspace_Update(t *testing.T) {
	tests := []struct {
		name string
		ws   *Workspace
		opts UpdateOptions
		want func(t *testing.T, got *Workspace)
	}{
		{
			name: "default",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name: internal.String("my-workspace"),
			},
			want: func(t *testing.T, got *Workspace) {
				assert.Equal(t, "my-workspace", got.Name)
			},
		},
		{
			name: "set trigger patterns",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				TriggerPatterns: []string{"/foo/**/*.tf"},
			},
			want: func(t *testing.T, got *Workspace) {
				assert.Equal(t, []string{"/foo/**/*.tf"}, got.TriggerPatterns)
			},
		},
		{
			name: "trigger patterns to tags regex",
			ws: &Workspace{
				Name:            "dev",
				Organization:    "acme",
				TriggerPatterns: []string{"/foo/**/*.tf"},
				Connection:      &Connection{},
			},
			opts: UpdateOptions{
				ConnectOptions: &ConnectOptions{
					TagsRegex: internal.String("\\d+"),
				},
			},
			want: func(t *testing.T, got *Workspace) {
				assert.Nil(t, got.TriggerPatterns)
				assert.Equal(t, "\\d+", got.Connection.TagsRegex)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.ws.Update(tt.opts)
			require.NoError(t, err)
			tt.want(t, tt.ws)
		})
	}
}

func TestWorkspace_UpdateConnection(t *testing.T) {
	tests := []struct {
		name string
		ws   *Workspace
		opts UpdateOptions
		want *bool
	}{
		{
			name: "connect",
			ws:   &Workspace{Name: "dev", Organization: "acme"},
			opts: UpdateOptions{
				Name: internal.String("my-workspace"),
				ConnectOptions: &ConnectOptions{
					RepoPath:      internal.String("leg100/otf"),
					VCSProviderID: internal.String("vcs-123"),
				},
			},
			want: internal.Bool(true),
		},
		{
			name: "disconnect",
			ws: &Workspace{
				Name:         "dev",
				Organization: "acme",
				Connection:   &Connection{},
			},
			opts: UpdateOptions{
				Name:       internal.String("my-workspace"),
				Disconnect: true,
			},
			want: internal.Bool(false),
		},
		{
			name: "modify connection",
			ws: &Workspace{
				Name:         "dev",
				Organization: "acme",
				Connection: &Connection{
					Repo:          "leg100/otf",
					VCSProviderID: "vcs-123",
				},
			},
			opts: UpdateOptions{
				Name: internal.String("my-workspace"),
				ConnectOptions: &ConnectOptions{
					RepoPath: internal.String("leg100/otf-demo"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ws.Update(tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
