package internal

import "fmt"

// SafeAppend appends strings to a slice whilst ensuring the slice is
// not modified.
//
// see: https://yourbasic.org/golang/gotcha-append/
func SafeAppend(a []string, b ...string) []string {
	dst := make([]string, len(a)+len(b))
	copy(dst, a)
	return append(dst, b...)
}

func ConvertSliceToString[S fmt.Stringer](src []S) []string {
	dst := make([]string, len(src))
	for i := range src {
		dst[i] = src[i].String()
	}
	return dst
}
