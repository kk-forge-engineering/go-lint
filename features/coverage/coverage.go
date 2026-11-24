// Package coverage holds what `go_lint cover` decides: which packages count toward the total, and the total itself.
package coverage

import (
	"io"
	"math"
	"path"
	"strings"

	"golang.org/x/tools/cover"
)

// Measured returns the packages whose statements count toward the total: every package but a test kit, the package
// `<name>_testing`. A test kit runs only under the tests, so its statements would count against the code it serves.
func (*UseCases) Measured(packages []string) []string {
	var measured []string
	for _, importPath := range packages {
		if !strings.HasSuffix(path.Base(importPath), "_testing") {
			measured = append(measured, importPath)
		}
	}
	return measured
}

// Total returns the percentage of statements the profile marks as run, rounded to a tenth as `go tool cover -func`
// prints its total, so a total it prints as 80.0% meets a floor of 80. A profile with no statements gives 0.
func (*UseCases) Total(profile io.Reader) (float64, error) {
	profiles, err := cover.ParseProfilesFromReader(profile)
	if err != nil {
		return 0, err
	}
	var statements, run int
	for _, file := range profiles {
		for _, block := range file.Blocks {
			statements += block.NumStmt
			if block.Count > 0 {
				run += block.NumStmt
			}
		}
	}
	if statements == 0 {
		return 0, nil
	}
	return math.Round(1000*float64(run)/float64(statements)) / 10, nil
}
