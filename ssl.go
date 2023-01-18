package otf

import "os"

const DefaultSSLCertsDir = "/etc/ssl/certs/ca-certificates.crt"

// SSLCertsDir returns the directory containing CA certificates.
func SSLCertsDir() string {
	// https://pkg.go.dev/crypto/x509#SystemCertPool
	if override, ok := os.LookupEnv("SSL_CERT_DIR"); ok {
		return override
	} else {
		return DefaultSSLCertsDir
	}
}
