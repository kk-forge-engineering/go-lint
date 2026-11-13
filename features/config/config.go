// Package config builds the golangci-lint configuration go-lint runs with: the set built into go-lint, plus the
// exclusions in the project's go-lint.yml.
package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"go.yaml.in/yaml/v3"
)

// fileName is the file at the Go module root that holds the project's exclusions.
const fileName = "go-lint.yml"

//go:embed golangci.yml
var base []byte

// exclusion is one entry of go-lint.yml: a golangci-lint exclusion rule and the reason the project gives for it.
type exclusion struct {
	Reason     string   `yaml:"reason,omitempty"`
	Path       string   `yaml:"path,omitempty"`
	PathExcept string   `yaml:"path-except,omitempty"`
	Linters    []string `yaml:"linters,omitempty"`
	Text       string   `yaml:"text,omitempty"`
	Source     string   `yaml:"source,omitempty"`
}

// For returns the configuration for the Go module at the root of module. It fails when the module
// keeps a golangci-lint configuration of its own, or when go-lint.yml holds an unknown key or an
// exclusion without a reason.
func For(module fs.FS) ([]byte, error) {
	for _, name := range []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"} {
		if _, err := fs.Stat(module, name); err == nil {
			return nil, fmt.Errorf("go-lint reads no %s: delete it, and put the project's exclusions in %s", name, fileName)
		}
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
	return yaml.Marshal(config)
}

// readExclusions returns the exclusions in go-lint.yml, or none when the file does not exist.
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
