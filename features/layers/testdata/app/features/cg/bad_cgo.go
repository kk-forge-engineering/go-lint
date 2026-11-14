// The cgo tool generates a file of this package in the build cache, outside the module.
package cg

// int one(void) { return 1; }
import "C"

import _ "example.com/app/infra/db" // want `features may not import infra \(example\.com/app/infra/db\)`

// One returns one. cgo's output of this file carries a generated header, and the surface check still reads it.
func One() int { return int(C.one()) } // want `One is a free function on the surface of features/cg`
