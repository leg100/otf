package internal

import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
)

// WebURL wraps the stdlib url.URL, restricting it to web URLs (i.e. those that
// use the http(s) scheme.
type WebURL struct {
	url.URL
}

func MustWebURL(rawURL string) *WebURL {
	u, err := NewWebURL(rawURL)
	if err != nil {
		panic(err.Error())
	}
	return u
}

// NewWebURL constructs a http(s) URL from a URL string. An error is returned if
// the string starts with a scheme other than http(s). If there is no scheme
// then the scheme is set to https.
func NewWebURL(rawURL string) (*WebURL, error) {
	scheme, _, hasScheme := strings.Cut(rawURL, "://")
	if hasScheme {
		if scheme != "https" && scheme != "http" {
			return nil, fmt.Errorf("cannot construct web url from non-http(s) scheme: %s", scheme)
		}
	} else {
		rawURL = "https://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	return &WebURL{URL: *u}, nil
}

// Type implements pflag.Value
func (*WebURL) Type() string { return "url" }

// Set implements pflag.Value
func (u *WebURL) Set(text string) error {
	newURL, err := NewWebURL(text)
	if err != nil {
		return err
	}
	*u = *newURL
	return nil
}

func (u *WebURL) Scan(text any) error {
	if text == nil {
		return nil
	}
	s, ok := text.(string)
	if !ok {
		return fmt.Errorf("expected database value to be a string: %#v", text)
	}
	return u.Set(s)
}

func (u *WebURL) Value() (driver.Value, error) {
	if u == nil {
		return nil, nil
	}
	return u.String(), nil
}

func (u WebURL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *WebURL) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	s := string(text)
	return u.Set(s)
}
