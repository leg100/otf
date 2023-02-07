package run

import (
	"time"

	"github.com/leg100/otf"
)

// phaseStatus is a mixin providing status functionality for a phase
type phaseStatus struct {
	status           otf.PhaseStatus
	statusTimestamps []otf.PhaseStatusTimestamp
}

func (p *phaseStatus) Status() otf.PhaseStatus                      { return p.status }
func (p *phaseStatus) StatusTimestamps() []otf.PhaseStatusTimestamp { return p.statusTimestamps }

func (p *phaseStatus) StatusTimestamp(status otf.PhaseStatus) (time.Time, error) {
	for _, rst := range p.statusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, ErrStatusTimestampNotFound
}

func (p *phaseStatus) updateStatus(status otf.PhaseStatus) {
	p.status = status
	p.statusTimestamps = append(p.statusTimestamps, otf.PhaseStatusTimestamp{
		Status:    status,
		Timestamp: otf.CurrentTimestamp(),
	})
}

func newPhaseStatus() *phaseStatus {
	p := &phaseStatus{}
	p.updateStatus(otf.PhasePending)
	return p
}
