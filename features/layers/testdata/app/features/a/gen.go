//go:build ignore

// The go generate tool runs this program. It is package main, so the check leaves it out.
package main

import (
	_ "example.com/app/entities/lone"
	_ "example.com/app/infra/db"
)

func main() {}
