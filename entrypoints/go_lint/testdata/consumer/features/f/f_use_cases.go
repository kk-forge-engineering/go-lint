package f

// UseCases is the feature's behaviour.
type UseCases struct{}

// Options holds the settings and dependencies of the use cases.
type Options struct{}

// New builds the use cases.
func New(Options) *UseCases { return &UseCases{} }
