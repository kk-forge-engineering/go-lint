package coverage

// UseCases picks the packages a coverage run measures, reads the total from a coverage profile and prints the
// `go test -json` stream of the run.
type UseCases struct{}

// Options holds the settings and dependencies of the use cases. The use cases need none yet.
type Options struct{}

// New returns the coverage use cases. It reads nothing from the options yet.
func New(Options) *UseCases { return &UseCases{} }
