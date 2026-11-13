package c

import "os"

// Clean removes the scratch file.
func Clean() {
	os.Remove("scratch") //nolint:errcheck
}
