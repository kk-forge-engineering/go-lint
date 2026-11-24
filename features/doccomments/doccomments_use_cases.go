package doccomments

import "golang.org/x/tools/go/analysis"

// UseCases gives out the doccomments analyzer.
type UseCases struct{}

// Options holds the settings and dependencies of the use cases. The analyzer needs none yet.
type Options struct{}

// New builds the use cases.
func New(Options) *UseCases { return &UseCases{} }

// Analyzer returns the doccomments analyzer.
func (*UseCases) Analyzer() *analysis.Analyzer { return analyzer }
