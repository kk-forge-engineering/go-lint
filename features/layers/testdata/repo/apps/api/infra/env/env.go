package env

import "os"

// Port returns the PORT environment variable, or "" when it is unset.
// An app's infra layer may read the environment.
func Port() string { return os.Getenv("PORT") }
