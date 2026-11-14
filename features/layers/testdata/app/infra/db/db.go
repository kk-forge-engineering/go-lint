package db

import (
	"time"

	_ "example.com/app/shared"
)

// Now reads the clock, which infra may do.
func Now() time.Time { return time.Now() }
