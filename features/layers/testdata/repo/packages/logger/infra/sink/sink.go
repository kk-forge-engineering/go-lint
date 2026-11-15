package sink

import (
	"os"
	"time"
)

// Level reads the environment, which a package's infra may not, and the clock, which it may.
func Level() (string, time.Time) {
	return os.Getenv("LOG_LEVEL"), time.Now() // want `os\.Getenv reads the environment, which a package's infra does not read`
}
