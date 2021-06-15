package ots

var _ Paginated = (*PaginatedMock)(nil)

type PaginatedMock struct {
	items   int
	current int
	size    int
	path    string
}

func NewPaginatedMock(path string, items, current, size int) *PaginatedMock {
	return &PaginatedMock{
		items:   items,
		current: current,
		size:    size,
		path:    path,
	}
}

func (p *PaginatedMock) GetItems() interface{} {
	return make([]int, p.items)
}

func (p *PaginatedMock) GetListOptions() ListOptions {
	return ListOptions{
		PageNumber: p.current,
		PageSize:   p.size,
	}
}

func (p *PaginatedMock) GetPath() string {
	return p.path
}
