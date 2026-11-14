// Package signature builds its use cases without the options.
package signature

type UseCases struct{}

func New() *UseCases { return &UseCases{} } // want `New takes the options and returns the use cases: func New\(o Options\) \*UseCases`
