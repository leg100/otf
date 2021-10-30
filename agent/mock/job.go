/*
Package mock provides mocks for the parent agent package
*/
package mock

import (
	"github.com/leg100/otf"
)

type Job struct {
	ID     string
	Status string
	DoFn   func(otf.Environment) error
}

func (j *Job) Do(exe otf.Environment) error {
	return j.DoFn(exe)
}

func (j *Job) GetID() string {
	return j.ID
}

func (j *Job) GetStatus() string {
	return j.Status
}
