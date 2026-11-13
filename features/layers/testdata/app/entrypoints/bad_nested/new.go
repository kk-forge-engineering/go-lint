package bad_nested

// Handle runs the app.
type Handle struct{}

// New returns the handle.
func New() Handle { return Handle{} }

// Run runs the handle.
func (Handle) Run() {}
