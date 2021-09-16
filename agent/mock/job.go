/*
Package mock provides mocks for the parent agent package
*/
package mock

import (
	"github.com/leg100/ots"
)

type Job struct {
	ID     string
	Status string
	DoFn   func(*ots.Environment) error
}

func (j *Job) Do(env *ots.Environment) error {
	return j.DoFn(env)
}

func (j *Job) GetID() string {
	return j.ID
}

func (j *Job) GetStatus() string {
	return j.Status
}
