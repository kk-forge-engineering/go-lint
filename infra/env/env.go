// Package env reads the process environment, which only infra does.
package env

import (
	"os"
	"slices"
	"strings"
)

// Without returns the process environment without the variable name.
func Without(name string) []string {
	return slices.DeleteFunc(os.Environ(), func(variable string) bool {
		return strings.HasPrefix(variable, name+"=")
	})
}
