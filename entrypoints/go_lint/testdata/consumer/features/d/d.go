package d

import "os"

// Clean removes the scratch file.
func (UseCases) Clean() {
	os.Remove("scratch") //nolint // the scratch file may be gone
}
