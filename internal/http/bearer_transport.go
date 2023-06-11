package http

import "net/http"

// HeadersTransport adds headers to each request
type HeadersTransport struct {
	rt    http.RoundTripper
	token string
}

func (bt *HeadersTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Bearer "+bt.token)
	r.Header.Set("Accept", "application/json")
	return bt.rt.RoundTrip(r)
}
