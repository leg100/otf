package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"
)

func TestMiddleware_AuthenticateToken(t *testing.T) {
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// implicitly responds with 200 OK
	})
	mw := AuthenticateToken(&fakeMiddlewareService{
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
	creds, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !ok {
		t.Skip("Export a valid GOOGLE_APPLICATION_CREDENTIALS before running this test")
	}

	ctx := context.Background()
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// implicitly responds with 200 OK
	})
	mw := AuthenticateToken(&fakeMiddlewareService{}, AuthenticateTokenConfig{
		GoogleJWTConfig: GoogleJWTConfig{
			Enabled:  true,
			Audience: "https://example.com",
		},
	})

	src, err := idtoken.NewTokenSource(ctx, "https://example.com", idtoken.WithCredentialsFile(creds))
	require.NoError(t, err)
	token, err := src.Token()
	require.NoError(t, err)
	r := httptest.NewRequest("GET", "/api/v2/runs/run-123", nil)
	token.SetAuthHeader(r)

	w := httptest.NewRecorder()
	mw(upstream).ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code, "output: %s", w.Body.String())
}

func Test_AuthenticateSession(t *testing.T) {
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// implicitly respond with 200 OK
	})
	mw := AuthenticateSession(&fakeMiddlewareService{
		sessionToken: "session.token",
	})

	t.Run("with session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/app/organizations", nil)
		r.AddCookie(&http.Cookie{Name: sessionCookie, Value: "session.token"})
		mw(upstream).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("without session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/app/organizations", nil)
		// deliberately omit session cookie
		mw(upstream).ServeHTTP(w, r)
		assert.Equal(t, 302, w.Code)
		loc, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Login(), loc.Path)
	})
}

type fakeMiddlewareService struct {
	agentToken    string
	registryToken string
	sessionToken  string
	userToken     string
}

func (f *fakeMiddlewareService) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	if f.agentToken == token {
		return nil, nil
	}
	return nil, errors.New("invalid")
}

func (f *fakeMiddlewareService) GetRegistrySession(ctx context.Context, token string) (*RegistrySession, error) {
	if f.registryToken == token {
		return nil, nil
	}
	return nil, errors.New("invalid")
}

func (f *fakeMiddlewareService) GetSession(ctx context.Context, token string) (*Session, error) {
	if f.sessionToken == token {
		return nil, nil
	}
	return nil, errors.New("invalid")
}

func (f *fakeMiddlewareService) GetUser(ctx context.Context, spec UserSpec) (*User, error) {
	if spec.AuthenticationToken != nil {
		if f.userToken == *spec.AuthenticationToken {
			return nil, nil
		}
	} else if spec.SessionToken != nil {
		if f.sessionToken == *spec.SessionToken {
			return nil, nil
		}
	} else if spec.Username != nil {
		// this is the google jwt check, so accept any username
		return nil, nil
	}
	return nil, errors.New("invalid")
}
