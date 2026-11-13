// The cgo tool generates a file of this package in the build cache, outside the module.
package cg

// int one(void) { return 1; }
import "C"

import _ "example.com/app/infra/db" // want `features may not import infra \(example\.com/app/infra/db\)`

// One returns one.
func One() int { return int(C.one()) }
