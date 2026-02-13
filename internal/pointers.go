package internal

//go:fix inline
func Ptr[T any](t T) *T { return new(t) }
