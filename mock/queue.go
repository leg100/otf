package mock

import (
	"github.com/leg100/otf"
)

var _ otf.Queue = (*Queue)(nil)

type Queue struct {
	Runs []*otf.Run
}

func (q *Queue) Add(run *otf.Run) error {
	q.Runs = append(q.Runs, run)

	return nil
}

func (q *Queue) Remove(run *otf.Run) error {
	for idx, r := range q.Runs {
		if run.ID == r.ID {
			q.Runs = append(q.Runs[:idx], q.Runs[idx+1:]...)
		}
	}

	return nil
}
