package main

import "os"

// Abstraction around methods that retrieve system directories
type Directories interface {
	UserHomeDir() (string, error)
}

// Wrapper around os.* funcs - expressly for the purpose substituting with fakes
// for testing purposes
type SystemDirectories struct{}

var _ Directories = (*SystemDirectories)(nil)

func (s *SystemDirectories) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}
