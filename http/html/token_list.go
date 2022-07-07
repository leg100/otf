package html

import "github.com/leg100/otf"

// tokenList exposes a list of tokens to a template
type tokenList struct {
	// list template expects pagination object but we don't paginate token
	// listing
	*otf.Pagination
	Items []*otf.Token
}

// sessionList exposes a list of sessions to a template
type sessionList struct {
	// list template expects pagination object but we don't paginate session
	// listing
	*otf.Pagination
	Items []*otf.Session
}
