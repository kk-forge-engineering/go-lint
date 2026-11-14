// Package misplaced keeps its use cases in a file not named for the feature.
package misplaced

type Options struct{}

type UseCases struct{} // want `UseCases goes in misplaced_use_cases\.go, the file named for its feature`

func New(o Options) *UseCases { return &UseCases{} } // want `New goes in misplaced_use_cases\.go, beside UseCases`
