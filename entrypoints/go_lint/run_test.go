package go_lint_test

import (
	"errors"
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// binary is go_lint, built once for the tests from cmd/go_lint.
var binary string

// goEnv holds the go tool's own cache folders.
var goEnv = map[string]string{}

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "go_lint_test_")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(dir, "go_lint")
	args := []string{"build", "-o", binary}
	// Under `mise run test:ci` the binary writes its coverage beside the tests', so the code it runs counts.
	flag.Parse()
	if coverDir := flag.Lookup("test.gocoverdir"); testing.CoverMode() != "" && coverDir.Value.String() != "" {
		args = append(args, "-cover", "-coverpkg=github.com/kk-forge-engineering/go-lint/v3/...")
		_ = os.Setenv("GOCOVERDIR", coverDir.Value.String())
	}
	build := exec.Command("go", append(args, "../../cmd/go_lint")...)
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic(err)
	}
	// The test reads the modules in testdata, so an edit to one runs the tests again. go test keys its cache only on
	// the files the test process opens, and go-lint reads a module in a process of its own.
	if err := readAll("testdata"); err != nil {
		panic(err)
	}
	for _, name := range []string{"GOCACHE", "GOMODCACHE"} {
		value, err := exec.Command("go", "env", name).Output()
		if err != nil {
			panic(err)
		}
		goEnv[name] = strings.TrimSpace(string(value))
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func readAll(dir string) error {
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		_, err = os.ReadFile(path)
		return err
	})
}

// goLint runs go-lint in testdata/consumer and returns its output and exit code.
func goLint(t *testing.T, args ...string) (string, int) {
	t.Helper()
	return goLintIn(t, "testdata/consumer", t.TempDir(), args...)
}

// goLintIn runs go-lint in dir, with home as the user's home folder, which holds the user's caches, golangci-lint's
// included.
func goLintIn(t *testing.T, dir, home string, args ...string) (string, int) {
	t.Helper()
	command := exec.Command(binary, args...)
	command.Dir = dir
	// The go tool's caches stay where they are, so the run compiles nothing again.
	command.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
		"GOCACHE="+goEnv["GOCACHE"], "GOMODCACHE="+goEnv["GOMODCACHE"], "GOLANGCI_LINT_CACHE="+filepath.Join(home, "golangci-lint"))
	output, err := command.CombinedOutput()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(output), exitErr.ExitCode()
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(output), 0
}

var issue = regexp.MustCompile(`(?m)^(\S+:\d+):\d+: .* \((\w+)\)$`)

func TestRunReportsTheAnalyzersAndAppliesTheProjectExclusions(t *testing.T) {
	output, code := goLint(t, "run", "./...")

	var got []string
	for _, match := range issue.FindAllStringSubmatch(output, -1) {
		got = append(got, match[1]+" "+match[2])
	}
	slices.Sort(got)
	// features/b's errcheck issue is excluded by testdata/consumer/go_lint.yml, and the request and TLS setting in features/f's
	// tests by the built-in exclusion for test files.
	want := []string{
		"features/a/a.go:3 layers", "features/c/c.go:7 nolintlint", "features/d/d.go:7 nolintlint",
		"features/f/f.go:13 noctx", "features/f/f.go:7 gochecknoinits", "features/f/tls.go:12 gosec",
		"features/f/tls.go:7 gosec", "features/g/g.go:7 naming",
	}
	if code != 1 || !slices.Equal(got, want) {
		t.Errorf("go_lint run exited %d with issues %v, want 1 with %v\n%s", code, got, want, output)
	}
}

// The fixture calls language.Parse of golang.org/x/text v0.3.0, which GO-2021-0113 names. The scan reads the
// vulnerability database online, so the test needs the network.
func TestVulnReportsAVulnerabilityTheCodeCalls(t *testing.T) {
	output, code := goLintIn(t, "testdata/vulnerable", t.TempDir(), "vuln")

	if code != 3 || !strings.Contains(output, "GO-2021-0113") {
		t.Errorf("go_lint vuln exited %d, want 3 with GO-2021-0113\n%s", code, output)
	}
}

// govulncheck takes the standard library's version from GOVERSION when the environment sets it, so a repo could
// set one past every fix and hide each finding there. go-lint asks the toolchain. The fixture calls url.Parse, and
// GO-2025-4010 there was fixed in go1.24.8, so a scan of go1.10 reports it and a scan of the toolchain does not.
func TestVulnIgnoresAGoVersionTheEnvironmentSets(t *testing.T) {
	t.Setenv("GOVERSION", "go1.10")

	output, _ := goLintIn(t, "testdata/vulnerable", t.TempDir(), "vuln")

	if !strings.Contains(output, "GO-2021-0113") || strings.Contains(output, "GO-2025-4010") {
		t.Errorf("go_lint vuln with GOVERSION=go1.10 output = %s\nwant GO-2021-0113 and no GO-2025-4010", output)
	}
}

// A configuration can load code, so go-lint writes its own on each run over a file it finds in its cache.
func TestRunReplacesAConfigurationFileFoundInItsCache(t *testing.T) {
	home := t.TempDir()
	goLintIn(t, "testdata/consumer", home, "run", "./...")
	files, err := filepath.Glob(filepath.Join(home, "*", "*", "go_lint", "*.yml"))
	if err != nil || len(files) == 0 {
		files, err = filepath.Glob(filepath.Join(home, "*", "go_lint", "*.yml"))
	}
	if err != nil || len(files) != 1 {
		t.Fatalf("configuration files in the cache = %v, %v; want one", files, err)
	}
	if err := os.WriteFile(files[0], []byte("version: \"2\"\nlinters:\n  default: none\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	output, code := goLintIn(t, "testdata/consumer", home, "run", "./...")

	if code != 1 || !strings.Contains(output, "(layers)") {
		t.Errorf("go_lint run over a planted configuration exited %d, want 1 with the layers issue\n%s", code, output)
	}
}

// golangci-lint keys the cached result of a package on its own files and its imports. The result of an entities
// package also depends on the packages that import it, so a second run must not reuse it.
func TestRunReportsAnEntityThatLostAConsumerSinceTheCachedRun(t *testing.T) {
	module, home := t.TempDir(), t.TempDir()
	write := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(module, path)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(module, path), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/shift\n\ngo 1.26.0\n")
	write("entities/user/user.go", "package user\n")
	write("features/a/a.go", "package a\n\nimport _ \"example.com/shift/entities/user\"\n")
	write("features/b/b.go", "package b\n\nimport _ \"example.com/shift/entities/user\"\n")
	if output, code := goLintIn(t, module, home, "run", "./..."); code != 0 {
		t.Fatalf("go-lint run with two consumers exited %d, want 0\n%s", code, output)
	}

	write("features/b/b.go", "package b\n")
	output, code := goLintIn(t, module, home, "run", "./...")

	if code != 1 || !strings.Contains(output, "only features/a imports entities/user") {
		t.Errorf("go-lint run after features/b dropped entities/user exited %d, want 1 with its report\n%s", code, output)
	}
}

func TestGoLintRefuses(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"an unknown command", []string{"linters"}, "unknown command"},
		{"a flag that changes the configuration", []string{"run", "--enable=lll"}, "--enable changes the built-in configuration"},
		{"another configuration file", []string{"run", "-c", "other.yml"}, "-c changes the built-in configuration"},
		{"a one-letter flag with its value attached", []string{"run", "-Derrcheck"}, "-D changes the built-in configuration"},
		{"a forbidden letter in a group of one-letter flags", []string{"run", "-vElll"}, "-E changes the built-in configuration"},
		{
			"a flag after a -- that is the value of the flag before it",
			[]string{"run", "--path-prefix", "--", "-D", "layers"},
			"-D changes the built-in configuration",
		},
		{"a flag that changes the exit code", []string{"run", "--issues-exit-code=0"}, "--issues-exit-code changes"},
		{"a filter to new issues", []string{"run", "-n"}, "-n changes the built-in configuration"},
		{"a flag to vuln", []string{"vuln", "-db", "file:///tmp/db"}, "-db is a flag, and go_lint vuln takes package patterns only"},
	} {
		t.Run(test.name, func(t *testing.T) {
			output, code := goLint(t, test.args...)
			if code != 2 || !strings.Contains(output, test.want) {
				t.Errorf("go_lint %v exited %d, want 2 with %q\n%s", test.args, code, test.want, output)
			}
		})
	}
}
