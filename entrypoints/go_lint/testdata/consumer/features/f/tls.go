package f

import "crypto/tls"

// Insecure returns a TLS configuration that trusts any certificate.
func (UseCases) Insecure() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}

// Silenced returns the same configuration under gosec's own comment, which go-lint ignores.
func (UseCases) Silenced() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true} // #nosec G402
}
