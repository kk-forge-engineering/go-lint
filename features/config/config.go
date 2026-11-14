// Package config builds the golangci-lint configuration go-lint runs with: the set built into go-lint, plus the
// module's path and the exclusions in the project's go_lint.yml.
package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"go.yaml.in/yaml/v3"
	"golang.org/x/mod/modfile"
)

// fileName is the file at the Go module root that holds the project's exclusions.
const fileName = "go_lint.yml"

//go:embed golangci.yml
var base []byte

// exclusion is one entry of go_lint.yml: a golangci-lint exclusion rule and the reason the project gives for it.
type exclusion struct {
	Reason     string   `yaml:"reason,omitempty"`
	Path       string   `yaml:"path,omitempty"`
	PathExcept string   `yaml:"path-except,omitempty"`
	Linters    []string `yaml:"linters,omitempty"`
	Text       string   `yaml:"text,omitempty"`
	Source     string   `yaml:"source,omitempty"`
}

// For returns the configuration for the Go module at the root of module. It sets importsKey on the layers linter.
// golangci-lint keys its cached results on the settings, so it drops them when the module's imports change. It fails
// on a go.mod module path that is missing or has no dot in its first element, on a .golangci file in the module, or
// on a bad go_lint.yml.
func (*UseCases) For(module fs.FS, importsKey string) ([]byte, error) {
	for _, name := range []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"} {
		if _, err := fs.Stat(module, name); err == nil {
			return nil, fmt.Errorf("go-lint reads no %s: delete it, and put the project's exclusions in %s", name, fileName)
		}
	}
	path, err := modulePath(module)
	if err != nil {
		return nil, err
	}
	exclusions, err := readExclusions(module)
	if err != nil {
		return nil, err
	}
	var config map[string]any
	if err := yaml.Unmarshal(base, &config); err != nil {
		return nil, fmt.Errorf("read the built-in configuration: %w", err)
	}
	linters := config["linters"].(map[string]any)
	rules := linters["exclusions"].(map[string]any)
	for _, rule := range exclusions {
		rule.Reason = ""
		rules["rules"] = append(rules["rules"].([]any), rule)
	}
	// gofumpt formats as `gofumpt` run in the module does: with the module's path, which go.mod gives the gofumpt
	// command.
	formatters := config["formatters"].(map[string]any)
	settings := formatters["settings"].(map[string]any)
	settings["gofumpt"] = map[string]any{"module-path": path}
	layers := linters["settings"].(map[string]any)["custom"].(map[string]any)["layers"].(map[string]any)
	layers["settings"] = map[string]any{"imports": importsKey}
	return yaml.Marshal(config)
}

// modulePath returns the module path in go.mod. It refuses a path without a dot in its first element, such as
// idea_packer: gofumpt takes an import of such a path for the standard library unless it starts like the module's
// own, and gci groups it apart, so where one such module imports another, go_lint run refuses what go_lint fmt
// writes.
func modulePath(module fs.FS) (string, error) {
	data, err := fs.ReadFile(module, "go.mod")
	if err != nil {
		return "", err
	}
	path := modfile.ModulePath(data)
	if path == "" {
		return "", errors.New("go.mod declares no module path")
	}
	if first, _, _ := strings.Cut(path, "/"); !strings.Contains(first, ".") {
		return "", fmt.Errorf("module path %s has no dot in its first element: name the module after its repository, "+
			"such as github.com/kk-forge-engineering/<repo>", path)
	}
	return path, nil
}

// readExclusions returns the exclusions in go_lint.yml, or none when the file does not exist.
func readExclusions(module fs.FS) ([]exclusion, error) {
	data, err := fs.ReadFile(module, fileName)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var project struct {
		Exclusions []exclusion `yaml:"exclusions"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&project); err != nil {
		return nil, fmt.Errorf("%s: %w", fileName, err)
	}
	for i, rule := range project.Exclusions {
		if strings.TrimSpace(rule.Reason) == "" {
			return nil, fmt.Errorf("%s: exclusion %d has no reason", fileName, i+1)
		}
	}
	return project.Exclusions, nil
}
