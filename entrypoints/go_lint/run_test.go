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
	"sync"
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
		args = append(args, "-cover", "-coverpkg=github.com/kk-forge-engineering/go-lint/v6/...")
		_ = os.Setenv("GOCOVERDIR", coverDir.Value.String())
	}
	build := exec.Command("go", append(args, "../../cmd/go_lint")...)
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
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

// readInputs reads, once per test process, every file a go_lint run depends on. go test keys its cache only on the
// files the test process opens, and only from m.Run on, while go_lint reads a module in a process of its own and is
// built by go build, which go test does not see. Reading the testdata modules and go_lint's own source here lets an
// edit to either run the tests again.
var readInputs = sync.OnceValue(func() error {
	for _, dir := range []string{"testdata", "../../features", "../../infra", "../../entrypoints", "../../cmd"} {
		if err := readAll(dir); err != nil {
			return err
		}
	}
	return nil
})

// goLintIn runs go-lint in dir, with home as the user's home folder, which holds the user's caches, golangci-lint's
// included.
func goLintIn(t *testing.T, dir, home string, args ...string) (string, int) {
	t.Helper()
	if err := readInputs(); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary, args...)
	command.Dir = dir
	// The go tool's caches stay where they are, so the run compiles nothing again.
	// golangci-lint refuses to start while another run holds its lock file in the temp folder. A temp folder under
	// home keeps the run apart from the go-lint that `mise run check` starts beside the tests.
	command.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME="+filepath.Join(home, ".cache"), "TMPDIR="+home,
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

func TestRunReportsGoLintsAnalyzersAndAppliesTheProjectExclusions(t *testing.T) {
	output, code := goLint(t, "run", "./...")

	var got []string
	for _, match := range issue.FindAllStringSubmatch(output, -1) {
		got = append(got, match[1]+" "+match[2])
	}
	slices.Sort(got)
	// The list leaves out features/b's errcheck issue, which testdata/consumer/go_lint.yml excludes, and the request
	// and TLS setting in features/f's tests, which the built-in exclusion for test files excludes.
	want := []string{
		"features/a/a.go:1 doccomments", "features/a/a.go:4 layers", "features/c/c.go:7 nolintlint", "features/d/d.go:7 nolintlint",
		"features/f/f.go:13 noctx", "features/f/f.go:7 gochecknoinits", "features/f/tls.go:12 gosec",
		"features/f/tls.go:7 gosec",
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

// A feature's infra may import only the entities its port imports. When the port drops one, the feature's own imports
// stay the same, so only the port's imports in the cache key make the second run report the infra's import.
func TestRunReportsAFeatureInfraImportThePortDroppedSinceTheCachedRun(t *testing.T) {
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
	write("features/a/port/port.go", "package port\n\nimport _ \"example.com/shift/entities/user\"\n")
	write("features/a/infra/infra.go", "package infra\n\nimport _ \"example.com/shift/entities/user\"\n")
	write("features/b/b.go", "package b\n\nimport _ \"example.com/shift/entities/user\"\n")
	if output, code := goLintIn(t, module, home, "run", "./..."); code != 0 {
		t.Fatalf("go-lint run with the port naming entities/user exited %d, want 0\n%s", code, output)
	}

	write("features/a/port/port.go", "package port\n")
	output, code := goLintIn(t, module, home, "run", "./...")

	if code != 1 || !strings.Contains(output, "a feature's infra imports only the entities its port features/a/port imports") {
		t.Errorf("go-lint run after the port dropped entities/user exited %d, want 1 with its report\n%s", code, output)
	}
}

// covered copies testdata/covered, where `go_lint cover` writes its profile, into a folder of the test's own.
func covered(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS("testdata/covered")); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestCoverLeavesOutTheTestKitsAndWritesTheProfile expects a total of 88.9%. The fixture's kit calc_testing holds
// most of its statements, and no test runs them: counted, they would bring the total down to 56.2%.
func TestCoverLeavesOutTheTestKitsAndWritesTheProfile(t *testing.T) {
	dir := covered(t)

	output, code := goLintIn(t, dir, t.TempDir(), "cover")

	if code != 0 || !strings.Contains(output, "coverage 88.9% of statements, floor 80%") {
		t.Errorf("go_lint cover exited %d, want 0 with a total of 88.9%%\n%s", code, output)
	}
	if _, err := os.Stat(filepath.Join(dir, "coverage", "coverage.out")); err != nil {
		t.Error(err)
	}
}

func TestCoverFailsUnderTheFloor(t *testing.T) {
	output, code := goLintIn(t, covered(t), t.TempDir(), "cover", "--min", "90")

	if code != 1 || !strings.Contains(output, "coverage 88.9% of statements is under the floor of 90%") {
		t.Errorf("go_lint cover --min 90 exited %d, want 1 with the total under the floor\n%s", code, output)
	}
}

func TestCoverFailsAndPrintsAFailedTest(t *testing.T) {
	dir := covered(t)
	failing := "package calc_test\n\nimport \"testing\"\n\nfunc TestFails(t *testing.T) { t.Error(\"the output of a test that fails\") }\n"
	if err := os.WriteFile(filepath.Join(dir, "calc", "fails_test.go"), []byte(failing), 0o600); err != nil {
		t.Fatal(err)
	}

	output, code := goLintIn(t, dir, t.TempDir(), "cover")

	if code != 1 || !strings.Contains(output, "the output of a test that fails") || strings.Contains(output, "floor") {
		t.Errorf("go_lint cover with a failing test exited %d, want 1 with the test's output and no total\n%s", code, output)
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
		{"a coverage floor under 80%", []string{"cover", "--min", "79.9"}, "--min 79.9 is outside 80 to 100"},
		{"a coverage floor that is no number", []string{"cover", "--min", "NaN"}, "--min NaN is outside 80 to 100"},
		{"a flag after a package pattern", []string{"cover", "./...", "--min", "90"}, "--min is a flag after a package pattern"},
		{"a flag cover does not take", []string{"cover", "-race"}, "flag provided but not defined: -race"},
	} {
		t.Run(test.name, func(t *testing.T) {
			output, code := goLint(t, test.args...)
			if code != 2 || !strings.Contains(output, test.want) {
				t.Errorf("go_lint %v exited %d, want 2 with %q\n%s", test.args, code, test.want, output)
			}
		})
	}
}
