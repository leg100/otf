package html

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
)

type Page[T any] struct {
	*resource.Page[T]
	*http.Request
}

// FirstItem retrieves the stringified ordinal number of the first item in the
// page.
func (p Page[T]) FirstItem() (string, error) {
	first, err := p.firstItem()
	if err != nil {
		return "", err
	}
	return strconv.Itoa(first), nil
}

func (p Page[T]) firstItem() (int, error) {
	if len(p.Items) == 0 {
		return 0, nil
	}
	var opts resource.PageOptions
	if err := decode.All(&opts, p.Request); err != nil {
		return 0, err
	}
	opts = opts.Normalize()
	return ((opts.PageNumber - 1) * opts.PageSize) + 1, nil
}

// LastItem retrieves the stringified ordinal number of the last item in the
// page.
func (p Page[T]) LastItem() (string, error) {
	first, err := p.firstItem()
	if err != nil {
		return "", err
	}
	last := max(0, first+len(p.Items)-1)
	return strconv.Itoa(last), nil
}

func (p Page[T]) PreviousPageLink() (string, error) {
	if p.PreviousPage == nil {
		return "", nil
	}
	u := p.URL.String()
	q := fmt.Sprintf("page[number]=%d", *p.PreviousPage)
	return mergeQuery(u, q)
}

func (p Page[T]) NextPageLink() (string, error) {
	if p.NextPage == nil {
		return "", nil
	}
	u := p.URL.String()
	q := fmt.Sprintf("page[number]=%d", *p.NextPage)
	return mergeQuery(u, q)
}
