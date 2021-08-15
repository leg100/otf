package mock

import (
	"github.com/leg100/ots"
)

var _ ots.Queue = (*Queue)(nil)

type Queue struct {
	Runs []*ots.Run
}

func (q *Queue) Add(run *ots.Run) error {
	q.Runs = append(q.Runs, run)

	return nil
}

func (q *Queue) Remove(run *ots.Run) error {
	for idx, r := range q.Runs {
		if run.ID == r.ID {
			q.Runs = append(q.Runs[:idx], q.Runs[idx+1:]...)
		}
	}

	return nil
}
