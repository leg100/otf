package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Encrypt plaintext using secret key. The returned string is
// base64-url-encoded.
func Encrypt(plaintext, secret []byte) (string, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// Prefix string with nonce
	return base64.URLEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

// Decrypt encrypted string using secret key. The encrypted string must be
// base64-url-encoded.
func Decrypt(encrypted string, secret []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	// Nonce is first 12 bytes, so decoding should at least be that length (plus
	// a multiple of 32 bytes for the ciphertext, but we'll let aesgcm.Open
	// check that).
	if len(decoded) < 12 {
		return nil, fmt.Errorf("size of decoded encrypted string is incorrect: %d", len(decoded))
	}

	return aesgcm.Open(nil, decoded[:12], decoded[12:], nil)
}
