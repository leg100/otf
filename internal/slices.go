package internal

// SafeAppend appends strings to a slice whilst ensuring the slice is
// not modified.
//
// see: https://yourbasic.org/golang/gotcha-append/
func SafeAppend(a []string, b ...string) []string {
	dst := make([]string, len(a)+len(b))
	copy(dst, a)
	return append(dst, b...)
}
