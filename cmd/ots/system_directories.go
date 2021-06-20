package main

import "os"

// Abstraction around methods that retrieve system directories - expressly for
// the purpose of substituting with fakes
type Directories interface {
	UserHomeDir() (string, error)
}

// Wrapper around os.* funcs
type SystemDirectories struct{}

var _ Directories = (*SystemDirectories)(nil)

func (s *SystemDirectories) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}
