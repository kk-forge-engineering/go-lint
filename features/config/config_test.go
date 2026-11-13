package config_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"go.yaml.in/yaml/v3"

	"github.com/kk-forge-engineering/go-lint/features/config"
)

func TestForAddsEachExclusionWithoutItsReason(t *testing.T) {
	module := fstest.MapFS{"go-lint.yml": {Data: []byte(`exclusions:
  - path: tools/
    linters: [errcheck]
    reason: the tools are throwaway scripts
`)}}

	data, err := config.For(module)
	if err != nil {
		t.Fatal(err)
	}

	var got struct {
		Linters struct {
			Exclusions struct {
				Rules []map[string]any `yaml:"rules"`
			} `yaml:"exclusions"`
		} `yaml:"linters"`
	}
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	// The project's rule comes after the built-in ones.
	rules := got.Linters.Exclusions.Rules
	last := rules[len(rules)-1]
	if last["path"] != "tools/" || last["reason"] != nil {
		t.Errorf("exclusion rules = %v, want the last one for tools/ without its reason", rules)
	}
}

func TestForFails(t *testing.T) {
	for _, test := range []struct {
		name, file, content, want string
	}{
		{
			"on an exclusion without a reason", "go-lint.yml", "exclusions:\n  - path: tools/\n    linters: [errcheck]\n",
			"go-lint.yml: exclusion 1 has no reason",
		},
		{
			"on a blank reason", "go-lint.yml", "exclusions:\n  - path: tools/\n    linters: [errcheck]\n    reason: \" \"\n",
			"go-lint.yml: exclusion 1 has no reason",
		},
		{
			"on a key it does not know", "go-lint.yml", "exclusions:\n  - path: tools/\n    reason: x\n    severity: low\n",
			"field severity not found",
		},
		{"on a setting other than exclusions", "go-lint.yml", "linters:\n  enable: [lll]\n", "field linters not found"},
		{
			"on a golangci-lint configuration of the module's own", ".golangci.yml", "version: \"2\"\n",
			"go-lint reads no .golangci.yml",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := config.For(fstest.MapFS{test.file: {Data: []byte(test.content)}})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Errorf("For() error = %v, want one containing %q", err, test.want)
			}
		})
	}
}

func TestForWithoutGoLintYmlKeepsTheBuiltInSet(t *testing.T) {
	data, err := config.For(fstest.MapFS{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "layers") {
		t.Errorf("configuration has no layers linter:\n%s", data)
	}
}
