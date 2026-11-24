package coverage

// UseCases picks the packages a coverage run measures and reads the total from a coverage profile.
type UseCases struct{}

// Options holds the settings and dependencies of the use cases. The use cases need none yet.
type Options struct{}

// New returns the coverage use cases. It reads nothing from the options yet.
func New(Options) *UseCases { return &UseCases{} }
