package internal

import (
	crypto "crypto/rand"
	"encoding/base64"
	"math/rand"
)

const alphanumeric = "abcdefghijkmnopqrstuvwxyzABCDEFGHIJKMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString generates a random string composed of alphanumeric
// characters of length size.
func GenerateRandomString(size int) string {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return string(buf)
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := crypto.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
