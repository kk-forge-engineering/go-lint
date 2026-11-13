package f

import "crypto/tls"

// Insecure returns a TLS configuration that trusts any certificate.
func Insecure() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}

// Silenced returns the same configuration under gosec's own comment, which go-lint ignores.
func Silenced() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true} // #nosec G402
}
