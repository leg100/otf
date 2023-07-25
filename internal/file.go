package internal

import "os"

// Exists checks whether a file or directory at the given path exists
func Exists(path string) bool {
	// Interpret any error from os.Stat as "not found"
	_, err := os.Stat(path)
	return err == nil
}
