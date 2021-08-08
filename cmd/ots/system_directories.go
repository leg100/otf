package main

import "os"

// Directories implementations provide access to system directories. Can be used
// for implementating a fake for testing purposes.
type Directories interface {
	UserHomeDir() (string, error)
}

// SystemDirectories implements Directories, wrapping os.* funcs
type SystemDirectories struct{}

var _ Directories = (*SystemDirectories)(nil)

func (s *SystemDirectories) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}
