package internal

import "strings"

func NewStringFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// SplitCSV splits a string with a comma delimited (a "comma-separated-value").
// It differs from strings.Split in that if no comma is found an empty slice is
// returned whereas strings.Split would return a single-element slice containing the
// original string.
func SplitCSV(csv string) []string {
	return strings.FieldsFunc(csv, func(r rune) bool { return r == ',' })
}

// FromStringCSV splits a comma-separated string into a slice of type T
func FromStringCSV[T ~string](csv string) (to []T) {
	from := SplitCSV(csv)
	to = make([]T, len(from))
	for i, f := range SplitCSV(csv) {
		to[i] = T(f)
	}
	return
}

func FromStringSlice[T ~string](from []string) (to []T) {
	to = make([]T, len(from))
	for i, f := range from {
		to[i] = T(f)
	}
	return
}

func ToStringSlice[T ~string](from []T) (to []string) {
	to = make([]string, len(from))
	for i, f := range from {
		to[i] = string(f)
	}
	return
}
