// Package port declares what features/a needs from outside.
package port

// The port names entities/user, so features/a/infra may import it.
import _ "example.com/app/entities/user"

// Store keeps values by key.
type Store interface{ Get(key string) string }
