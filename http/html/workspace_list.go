package html

import (
	"net/url"
	"strconv"

	"github.com/leg100/otf"
)

type WorkspaceList struct {
	*otf.WorkspaceList
}

// NextPagePath returns a URL path for the next page of results for a workspace
// listing.
func (l *WorkspaceList) NextPagePath() *string {
	if len(l.Items) == 0 {
		return nil
	}
	u := url.URL{
		Path: listWorkspacePath(l.Items[0]),
	}
	q := url.Values{}
	q.Add("page[number]", strconv.Itoa(l.NextPage))
	q.Add("page[size]", strconv.Itoa(otf.DefaultPageSize))
	u.RawQuery = q.Encode()

	return otf.String(u.String())
}
