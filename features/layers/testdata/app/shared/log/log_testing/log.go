// Package log_testing is the test kit of shared/log. It may set up the clock as a test does, but it follows the layer of
// the package it serves, so it imports no entity.
package log_testing

import (
	"time"

	_ "example.com/app/entities/user" // want `shared may not import entities \(example\.com/app/entities/user\)`
)

// Now reads the clock.
func Now() time.Time { return time.Now() }
