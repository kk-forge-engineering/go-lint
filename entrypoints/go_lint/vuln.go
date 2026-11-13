package go_lint

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/vuln/scan"
)

// vuln runs govulncheck over the packages, ./... when none is named, and returns its exit code, which is 3 when the
// code calls a known vulnerability. It takes no flag: a flag could choose another database, count fewer findings or
// turn a finding into a pass.
func vuln(patterns []string) int {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "-") {
			fmt.Fprintf(os.Stderr, "go_lint: %s is a flag, and go_lint vuln takes package patterns only\n", pattern)
			return 2
		}
	}
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	command := scan.Command(context.Background(), patterns...)
	// govulncheck takes the standard library's version from GOVERSION when the environment sets it, so a version past
	// every fix would hide each finding there. Without it, govulncheck asks the toolchain.
	command.Env = slices.DeleteFunc(os.Environ(), func(variable string) bool {
		return strings.HasPrefix(variable, "GOVERSION=")
	})
	err := command.Start()
	if err == nil {
		err = command.Wait()
	}
	if err == nil {
		return 0
	}
	fmt.Fprintln(os.Stderr, "go_lint:", err)
	var exit interface{ ExitCode() int }
	if errors.As(err, &exit) {
		return exit.ExitCode()
	}
	return 1
}
