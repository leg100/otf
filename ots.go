package ots

import (
	"math/rand"
	"time"
)

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaxPageSize       = 100

	DefaultUserID   = "user-123"
	DefaultUsername = "ots"

	alphanumeric = "abcdefghijkmnopqrstuvwxyzABCDEFGHIJKMNOPQRSTUVWXYZ0123456789"
)

func String(str string) *string { return &str }
func Int(i int) *int            { return &i }
func Int64(i int64) *int64      { return &i }
func UInt(i uint) *uint         { return &i }

func GenerateRandomString(size int) string {
	// Without this, Go would generate the same random sequence each run.
	rand.Seed(time.Now().UnixNano())

	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return string(buf)
}
