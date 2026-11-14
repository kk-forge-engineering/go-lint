package b

import "os"

// Clean removes the scratch file.
func (UseCases) Clean() {
	os.Remove("scratch")
}
