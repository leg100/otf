package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"
)

func TestMiddleware_AuthenticateToken(t *testing.T) {
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// implicitly responds with 200 OK
	})
	mw := NewAuthTokenMiddleware(&fakeMiddlewareService{
		agentToken:    "agent.token",
		registryToken: "registry.token",
		userToken:     "user.token",
	}, AuthenticateTokenConfig{SiteToken: "site.token"})

	tests := []struct {
		name string
		// add bearer token to http request; nil omits the token
		token *string
		want  int
	}{
		{
			name:  "valid user token",
			token: otf.String("user.token"),
			want:  http.StatusOK,
		},
		{
			name:  "valid site token",
			token: otf.String("site.token"),
			want:  http.StatusOK,
		},
		{
			name:  "valid agent token",
			token: otf.String("agent.token"),
			want:  http.StatusOK,
		},
		{
			name:  "invalid token",
			token: otf.String("invalidToken"),
			want:  http.StatusUnauthorized,
		},
		{
			name:  "malformed token",
			token: otf.String("malfo rmedto ken"),
			want:  http.StatusUnauthorized,
		},
		{
			name: "missing token",
			want: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/api/v2/runs/run-123", nil)
			if tt.token != nil {
				r.Header.Add("Authorization", "Bearer "+*tt.token)
			}
			mw(upstream).ServeHTTP(w, r)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestMiddleware_AuthenticateToken_GoogleJWT(t *testing.T) {
	credspath := getGoogleCredentialsPath(t)

	ctx := context.Background()
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// implicitly responds with 200 OK
	})

	tests := []struct {
		name string
		cfg  GoogleJWTConfig
		want int
	}{
		{
			name: "disabled",
			want: http.StatusUnauthorized,
		},
		{
			name: "enabled",
			cfg:  GoogleJWTConfig{Enabled: true},
			want: http.StatusOK,
		},
		{
			name: "validate valid audience",
			cfg:  GoogleJWTConfig{Enabled: true, Audience: "https://example.com"},
			want: http.StatusOK,
		},
		{
			name: "validate invalid audience",
			cfg:  GoogleJWTConfig{Enabled: true, Audience: "http://l33th4cks.io"},
			want: http.StatusUnauthorized,
		},
		{
			name: "custom header",
			cfg:  GoogleJWTConfig{Enabled: true, Header: "x-goog-iap-jwt-assertion"},
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewAuthTokenMiddleware(&fakeMiddlewareService{}, AuthenticateTokenConfig{
				GoogleJWTConfig: tt.cfg,
			})

			src, err := idtoken.NewTokenSource(ctx, "https://example.com", idtoken.WithCredentialsFile(credspath))
			require.NoError(t, err)
			token, err := src.Token()
			require.NoError(t, err)
			r := httptest.NewRequest("GET", "/api/v2/runs/run-123", nil)
			if tt.cfg.Header == "" {
				token.SetAuthHeader(r) // uses standard "Authentication: bearer <>" header
			} else {
				r.Header.Set(tt.cfg.Header, token.AccessToken)
			}

			w := httptest.NewRecorder()
			mw(upstream).ServeHTTP(w, r)
			assert.Equal(t, tt.want, w.Code, "output: %s", w.Body.String())
		})
	}
}
