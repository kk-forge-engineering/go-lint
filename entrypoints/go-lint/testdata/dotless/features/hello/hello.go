package hello

import (
	"dotless/app/shared/greet"
	"fmt"
)

// Print prints a greeting for name.
func Print(name string) { fmt.Println(greet.Greeting(name)) }
