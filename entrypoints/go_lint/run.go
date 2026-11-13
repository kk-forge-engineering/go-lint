// Package go_lint is go-lint's command: golangci-lint with the layers analyzer and the configuration built into
// go-lint, and govulncheck.
package go_lint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/golangci/golangci-lint/v2/pkg/commands"
	"github.com/golangci/golangci-lint/v2/pkg/exitcodes"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	module "github.com/kk-forge-engineering/go-lint/v2"
	"github.com/kk-forge-engineering/go-lint/v2/features/config"
	"github.com/kk-forge-engineering/go-lint/v2/features/layers"
)

const usage = `usage: go_lint run [packages]   lint, and check the formatting
       go_lint fmt [paths]      format
       go_lint vuln [packages]  report the known vulnerabilities the code calls`

// forbidden lists the golangci-lint flags that would change the built-in configuration, so a run on a laptop and a
// run in CI check the same rules.
// The list also holds the flags that decide whether a run fails: the exit code, the test files, the issue limits and
// the filters to new issues.
var forbidden = []string{
	"-c", "--config", "--no-config", "-E", "--enable", "-D", "--disable", "--enable-only", "--default", "--fast-only",
	"--issues-exit-code", "--tests", "--max-issues-per-linter", "--max-same-issues", "--uniq-by-line",
	"-n", "--new", "--new-from-rev", "--new-from-patch", "--new-from-merge-base", "--whole-files",
}

// Run runs go-lint with the process's arguments, and exits with the exit code of golangci-lint or govulncheck.
func Run() {
	if len(os.Args) > 1 && os.Args[1] == "vuln" {
		os.Exit(vuln(os.Args[2:]))
	}
	if err := prepare(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "go_lint:", err)
		os.Exit(2)
	}
	// golangci-lint prints an error of its command line, but not one from reading --color or --verbose, so go-lint
	// prints each error, as golangci-lint's own main does.
	if err := commands.Execute(commands.BuildInfo{Version: cacheVersion(), GoVersion: "", Date: "", Commit: ""}); err != nil {
		fmt.Fprintln(os.Stderr, "go_lint:", err)
		os.Exit(exitcodes.Failure)
	}
}

// cacheVersion returns the version golangci-lint keys its cache of results on. The version is
// `module.Version` only for a build of that release's tag. golangci-lint keys any other build on
// the binary's own hash, so a change to go-lint between releases never reuses the results of
// another build.
func cacheVersion() string {
	if info, isAvailable := debug.ReadBuildInfo(); isAvailable && info.Main.Version == "v"+module.Version {
		return module.Version
	}
	return "(devel)"
}

// prepare checks the arguments, writes the configuration and sets the arguments golangci-lint reads.
func prepare(args []string) error {
	if len(args) == 0 || (args[0] != "run" && args[0] != "fmt") {
		return fmt.Errorf("unknown command\n%s", usage)
	}
	// A `--` may be the value of the flag before it, so the check reads every argument. A package pattern or a path
	// never starts with a dash.
	for _, arg := range args[1:] {
		if flag := forbiddenFlag(arg); flag != "" {
			return fmt.Errorf("%s changes the built-in configuration, so go-lint takes no such flag", flag)
		}
	}
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	data, err := config.For(os.DirFS(root))
	if err != nil {
		return err
	}
	file, err := writeConfig(data)
	if err != nil {
		return err
	}
	register.Plugin("layers", func(any) (register.LinterPlugin, error) { return layersPlugin{}, nil })
	os.Args = slices.Concat([]string{os.Args[0], args[0], "--config", file}, args[1:])
	return nil
}

// forbiddenFlag returns the forbidden flag an argument sets, or "". It reads a group of one-letter flags such as
// `-vElll` letter by letter, up to `-j`, the one allowed letter that takes the rest of the group as its value.
func forbiddenFlag(arg string) string {
	name, _, _ := strings.Cut(arg, "=")
	if slices.Contains(forbidden, name) {
		return name
	}
	if strings.HasPrefix(arg, "--") || !strings.HasPrefix(arg, "-") {
		return ""
	}
	for _, letter := range arg[1:] {
		flag := "-" + string(letter)
		if slices.Contains(forbidden, flag) {
			return flag
		}
		if letter == 'j' {
			return ""
		}
	}
	return ""
}

// writeConfig writes the configuration into the user's cache, named by its content, and returns the file's path.
// golangci-lint ends the process when it is done, so go-lint could not delete a file per run. One file per
// configuration stays instead. Each run writes it afresh, through a new file renamed over it: a file found under that
// name may be one someone else put there, and a configuration can load code.
func writeConfig(data []byte) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cache, "go_lint")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	temporary, err := os.CreateTemp(dir, "*.tmp")
	if err != nil {
		return "", err
	}
	_, err = temporary.Write(data)
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	file := filepath.Join(dir, hex.EncodeToString(sum[:8])+".yml")
	return file, os.Rename(temporary.Name(), file)
}

// moduleRoot returns the folder of the nearest go.mod in the working directory or above it.
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("no go.mod above the working directory")
		}
		dir = filepath.Dir(dir)
	}
}

type layersPlugin struct{}

func (layersPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{layers.Analyzer}, nil
}

// GetLoadMode asks for type information, which gives the analyzer the names of imported packages.
func (layersPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
