package config

// UseCases writes the golangci-lint configuration go-lint runs with.
type UseCases struct{}

// Options holds the settings and dependencies of the use cases. The configuration needs none yet.
type Options struct{}

// New builds the use cases.
func New(Options) *UseCases { return &UseCases{} }
