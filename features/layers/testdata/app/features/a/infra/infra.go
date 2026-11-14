// Package infra holds the adapters only features/a needs. It may read the clock.
package infra

import (
	"time"

	_ "example.com/app/entities/user"
	_ "example.com/app/features/a/port"
	_ "example.com/app/infra/db"
	_ "example.com/app/shared"
)

// Now reads the clock.
func Now() time.Time { return time.Now() }
