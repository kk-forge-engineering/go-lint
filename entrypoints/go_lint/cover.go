package go_lint

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kk-forge-engineering/go-lint/v6/features/coverage"
)

// coverageProfile is the file `go_lint cover` writes, in the folder the KK Forge testing standard keeps a package's
// reports in. SonarQube reads it as `sonar.go.coverage.reportPaths`.
const coverageProfile = "coverage/coverage.out"

// floor is the least total the KK Forge testing standard allows. A project may raise it.
const floor = 80

// cover runs go test over the packages, ./... when none is named, and measures every package but a test kit.
// It returns go test's exit code when a test fails, 1 when the total is under --min, and 2 on a usage error or when
// nothing was measured.
func cover(args []string) int {
	flags := flag.NewFlagSet("go_lint cover", flag.ContinueOnError)
	least := flags.Float64("min", floor, fmt.Sprintf("the least total coverage, in percent, from %d to 100", floor))
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if !(*least >= floor && *least <= 100) {
		fmt.Fprintf(os.Stderr, "go_lint: --min %g is outside %d to 100: the floor is %d%%, and a project may raise it\n", *least, floor, floor)
		return 2
	}
	patterns := flags.Args()
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "-") {
			fmt.Fprintf(os.Stderr, "go_lint: %s is a flag after a package pattern: put --min first\n", pattern)
			return 2
		}
	}
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	useCases := coverage.New(coverage.Options{})
	total, err := coverTests(useCases, patterns)
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "go_lint:", err)
		return 2
	}
	if total < *least {
		fmt.Fprintf(os.Stderr, "go_lint: coverage %.1f%% of statements is under the floor of %g%%\n", total, *least)
		return 1
	}
	fmt.Printf("coverage %.1f%% of statements, floor %g%%\n", total, *least)
	return 0
}

// coverTests runs go test over the patterns, writes the profile and returns its total. It returns go test's
// *exec.ExitError when a test fails.
func coverTests(useCases *coverage.UseCases, patterns []string) (float64, error) {
	list := exec.CommandContext(context.Background(), "go", append([]string{"list"}, patterns...)...)
	list.Stderr = os.Stderr
	listed, err := list.Output()
	if err != nil {
		// The error keeps go list's exit code out of reach, so a failed go list exits 2, as a run that measured nothing.
		return 0, fmt.Errorf("go list: %v", err)
	}
	measured := useCases.Measured(strings.Fields(string(listed)))
	if len(measured) == 0 {
		return 0, errors.New("every package is a test kit, so there is no code to measure")
	}
	if err := os.MkdirAll(filepath.Dir(coverageProfile), 0o755); err != nil {
		return 0, err
	}
	args := append([]string{"test", "-coverpkg=" + strings.Join(measured, ","), "-coverprofile=" + coverageProfile}, patterns...)
	test := exec.CommandContext(context.Background(), "go", args...)
	test.Stdout, test.Stderr = os.Stdout, os.Stderr
	if err := test.Run(); err != nil {
		return 0, err
	}
	profile, err := os.ReadFile(coverageProfile)
	if err != nil {
		return 0, err
	}
	return useCases.Total(bytes.NewReader(profile))
}
