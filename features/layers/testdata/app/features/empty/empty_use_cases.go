// Package empty is a fixture whose use cases file declares Options but not UseCases or New.
package empty // want `empty_use_cases\.go holds type UseCases and func New\(o Options\) \*UseCases`

type Options struct{}
