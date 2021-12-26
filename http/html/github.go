package html

import (
	"context"
	"net/url"
	gourl "net/url"

	gogithub "github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"
)

const (
	githubCallbackPath       = "/github/callback"
	DefaultGithubRedirectURL = "https://localhost" + githubCallbackPath
)

var githubScopes = []string{"user:email", "read:org"}

type GithubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Hostname     string
}

// githubOAuthApp is a github oauth app, responsible for handling authentication
// via oauth.
type githubOAuthApp struct {
	*oauth
	hostname string
}

func newGithubOAuthApp(config GithubConfig) (*githubOAuthApp, error) {
	endpoint, err := oauthEndpoint(config.Hostname)
	if err != nil {
		return nil, err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     endpoint,
		Scopes:       githubScopes,
	}

	gh := githubOAuthApp{
		hostname: config.Hostname,
		oauth: &oauth{
			Config: oauthConfig,
		},
	}

	return &gh, nil
}

// Build a new client: using hostname, oauth config, and token.
func (a *githubOAuthApp) newClient(ctx context.Context, token *oauth2.Token) (*gogithub.Client, error) {
	httpClient := a.Config.Client(ctx, token)

	if isGithubEnterprise(a.hostname) {
		return gogithub.NewEnterpriseClient(enterpriseBaseURL(a.hostname), enterpriseUploadURL(a.hostname), httpClient)
	}

	return gogithub.NewClient(httpClient), nil
}

func oauthEndpoint(hostname string) (oauth2.Endpoint, error) {
	if !isGithubEnterprise(hostname) {
		return githubOAuth2.Endpoint, nil
	}

	tokenEndpoint, err := replaceHost(githubOAuth2.Endpoint.TokenURL, hostname)
	if err != nil {
		return oauth2.Endpoint{}, err
	}

	authEndpoint, err := replaceHost(githubOAuth2.Endpoint.AuthURL, hostname)
	if err != nil {
		return oauth2.Endpoint{}, err
	}

	return oauth2.Endpoint{
		TokenURL: tokenEndpoint,
		AuthURL:  authEndpoint,
	}, nil
}

// Return a github enterprise URL from a hostname
func enterpriseBaseURL(hostname string) string   { return enterpriseURL(hostname, "/api/v3") }
func enterpriseUploadURL(hostname string) string { return enterpriseURL(hostname, "/api/uploads") }
func enterpriseURL(hostname, path string) string {
	return (&url.URL{Scheme: "https", Host: hostname, Path: path}).String()
}

func isGithubEnterprise(hostname string) bool {
	return hostname != "github.com"
}

func replaceHost(url string, newHost string) (string, error) {
	u, err := gourl.Parse(url)
	if err != nil {
		return "", err
	}

	u.Host = newHost

	return u.String(), nil
}
