package app

import (
	"github.com/leg100/otf"
)

var _ otf.ApplyService = (*ApplyService)(nil)

type ApplyService struct {
	db otf.RunStore
}

func NewApplyService(db otf.RunStore) *ApplyService {
	return &ApplyService{
		db: db,
	}
}

func (s ApplyService) Get(id string) (*otf.Apply, error) {
	run, err := s.db.Get(otf.RunGetOptions{ApplyID: &id})
	if err != nil {
		return nil, err
	}
	return run.Apply, nil
}
