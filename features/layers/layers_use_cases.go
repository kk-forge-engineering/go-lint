package layers

import "golang.org/x/tools/go/analysis"

// UseCases gives out the layers analyzer and the key of a module's imports that golangci-lint caches its results on.
type UseCases struct{}

// Options holds the settings and dependencies of the use cases. The analyzer needs none yet.
type Options struct{}

// New builds the use cases.
func New(Options) *UseCases { return &UseCases{} }

// Analyzer returns the layers analyzer.
func (*UseCases) Analyzer() *analysis.Analyzer { return analyzer }
