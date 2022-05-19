package inmem

import (
	"context"
	"errors"

	"github.com/leg100/otf"
)

type testCache struct {
	cache map[string][]byte

	otf.Cache
}

func newTestCache() *testCache { return &testCache{cache: make(map[string][]byte)} }

func (c *testCache) Set(key string, val []byte) error {
	c.cache[key] = val

	return nil
}

func (c *testCache) Get(key string) ([]byte, error) {
	val, ok := c.cache[key]
	if !ok {
		return nil, errors.New("not found")
	}

	return val, nil
}

func (c *testCache) AppendChunk(key string, chunk []byte) error {
	c.cache[key] = append(c.cache[key], chunk...)

	return nil
}

type testChunkStore struct {
	store map[string]otf.Chunk

	otf.ChunkStore
}

func newTestChunkStore() *testChunkStore { return &testChunkStore{store: make(map[string]otf.Chunk)} }

func (s *testChunkStore) GetChunk(ctx context.Context, id string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	if opts.Limit == 0 {
		return s.store[id].Cut(opts)
	}
	return s.store[id].Cut(opts)
}

func (s *testChunkStore) PutChunk(ctx context.Context, id string, chunk otf.Chunk) error {
	if val, ok := s.store[id]; ok {
		s.store[id] = val.Append(chunk)
	} else {
		s.store[id] = chunk
	}

	return nil
}

type fakeWorkspaceService struct {
	workspaces []*otf.Workspace
	otf.WorkspaceService
}

func (s *fakeWorkspaceService) List(_ context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items: s.workspaces,
	}, nil
}

type fakeRunService struct {
	runs []*otf.Run
	otf.RunService
}

func (s *fakeRunService) List(_ context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	var items []*otf.Run
	for _, r := range s.runs {
		if *opts.WorkspaceID != r.Workspace.ID {
			continue
		}
		// if statuses are specified then run must match one of them.
		if len(opts.Statuses) > 0 && !otf.ContainsRunStatus(opts.Statuses, r.Status) {
			continue
		}
		items = append(items, r)
	}
	return &otf.RunList{
		Items: items,
	}, nil
}

var _ otf.Queue = (*fakeQueue)(nil)

type fakeQueue struct {
	Runs []*otf.Run
}

func (q *fakeQueue) Update(run *otf.Run) *otf.Run {
	q.Runs = append(q.Runs, run)
	return nil
}

func (q *fakeQueue) Len() int {
	return len(q.Runs)
}
