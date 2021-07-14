package app

import (
	"github.com/leg100/ots"
)

var _ ots.ApplyService = (*ApplyService)(nil)

type ApplyService struct {
	db ots.RunStore
}

func NewApplyService(db ots.RunStore) *ApplyService {
	return &ApplyService{
		db: db,
	}
}

func (s ApplyService) Get(id string) (*ots.Apply, error) {
	run, err := s.db.Get(ots.RunGetOptions{ApplyID: &id})
	if err != nil {
		return nil, err
	}
	return run.Apply, nil
}
