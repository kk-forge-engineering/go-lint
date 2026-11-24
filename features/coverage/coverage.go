// Package coverage holds what `go_lint cover` decides: which packages count toward the total, the total itself, and
// what of a `go test -json` stream a person reads.
package coverage

import (
	"bufio"
	"encoding/json"
	"io"
	"maps"
	"math"
	"path"
	"slices"
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

// event is one line of `go test -json`.
type event struct {
	Action  string
	Package string
	Test    string
	Output  string
}

// Print writes what `go test` prints without -json: the output of each package and each build, and the output of a
// failed test. It drops the output of a test that passed or was skipped, which -json carries as -v would. A line that
// is no event, such as a message of the go tool, passes through as it is.
func (*UseCases) Print(stream io.Reader, out io.Writer) error {
	tests := map[[2]string]*strings.Builder{}
	reader := bufio.NewReader(stream)
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			if err := printLine(line, tests, out); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

// printLine prints one line of the stream, or keeps a test's output in tests until the test ends.
func printLine(line []byte, tests map[[2]string]*strings.Builder, out io.Writer) error {
	var e event
	if json.Unmarshal(line, &e) != nil {
		_, err := out.Write(line)
		return err
	}
	if e.Test == "" {
		// The loop prints the output still kept for the package's tests when the package's own event arrives. A test that
		// times out or calls os.Exit gets no fail event, so its output, the timeout's panic included, would be lost. go test
		// sends the package's event after its tests end, so any output still kept then belongs to such a test.
		for _, key := range slices.SortedFunc(maps.Keys(tests), func(a, b [2]string) int { return strings.Compare(a[1], b[1]) }) {
			if key[0] == e.Package {
				if _, err := io.WriteString(out, tests[key].String()); err != nil {
					return err
				}
				delete(tests, key)
			}
		}
		_, err := io.WriteString(out, e.Output)
		return err
	}
	key := [2]string{e.Package, e.Test}
	switch e.Action {
	case "output":
		if tests[key] == nil {
			tests[key] = &strings.Builder{}
		}
		tests[key].WriteString(e.Output)
	case "fail":
		if output := tests[key]; output != nil {
			if _, err := io.WriteString(out, output.String()); err != nil {
				return err
			}
		}
		delete(tests, key)
	case "pass", "skip":
		delete(tests, key)
	}
	return nil
}
