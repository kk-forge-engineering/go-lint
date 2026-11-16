package env

import "os"

// Level returns the value of the `LOG_LEVEL` environment variable.
// The layers check lets infra read the environment here, because this module sits outside the `apps/` and `packages/` of a go.work.
func Level() string { return os.Getenv("LOG_LEVEL") }
