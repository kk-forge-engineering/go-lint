package infra

import "os"

// Format reads the environment, which a feature's infra in a package may not.
func Format() (string, bool) {
	return os.LookupEnv("LOG_FORMAT") // want `os\.LookupEnv reads the environment, which a package's infra does not read`
}
