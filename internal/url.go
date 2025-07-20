package internal

import "net/url"

// URL wraps the stdlib url.URL
type URL struct {
	*url.URL
}

func MustURL(rawURL string) *URL {
	u, err := NewURL(rawURL)
	if err != nil {
		panic(err.Error())
	}
	return u
}

func NewURL(rawURL string) (*URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	return &URL{URL: u}, nil
}

// Type implements pflag.Value
func (*URL) Type() string { return "url" }

// Set implements pflag.Value
func (u *URL) Set(text string) error {
	newURL, err := NewURL(text)
	if err != nil {
		return err
	}
	*u = *newURL
	return nil
}
