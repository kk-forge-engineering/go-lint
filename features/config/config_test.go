package config_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"go.yaml.in/yaml/v3"

	"github.com/kk-forge-engineering/go-lint/v4/features/config"
)

// goMod is the go.mod of the modules the tests build their configuration for.
var goMod = &fstest.MapFile{Data: []byte("module github.com/kk-forge-engineering/idea-packer/backend\n\ngo 1.26.0\n")}

func TestForAddsEachExclusionWithoutItsReason(t *testing.T) {
	module := fstest.MapFS{"go.mod": goMod, "go_lint.yml": {Data: []byte(`exclusions:
  - path: tools/
    linters: [errcheck]
    reason: the tools are throwaway scripts
`)}}

	data, err := config.New(config.Options{}).For(module, "key")
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
			"on an exclusion without a reason", "go_lint.yml", "exclusions:\n  - path: tools/\n    linters: [errcheck]\n",
			"go_lint.yml: exclusion 1 has no reason",
		},
		{
			"on a blank reason", "go_lint.yml", "exclusions:\n  - path: tools/\n    linters: [errcheck]\n    reason: \" \"\n",
			"go_lint.yml: exclusion 1 has no reason",
		},
		{
			"on a key it does not know", "go_lint.yml", "exclusions:\n  - path: tools/\n    reason: x\n    severity: low\n",
			"field severity not found",
		},
		{"on a setting other than exclusions", "go_lint.yml", "linters:\n  enable: [lll]\n", "field linters not found"},
		{
			"on a golangci-lint configuration of the module's own", ".golangci.yml", "version: \"2\"\n",
			"go-lint reads no .golangci.yml",
		},
		{"on a go.mod without a module path", "go.mod", "go 1.26.0\n", "go.mod declares no module path"},
		{
			"on a module path without a dot in its first element", "go.mod", "module idea_packer/backend\n",
			"module path idea_packer/backend has no dot in its first element: name the module after its repository, " +
				"such as github.com/kk-forge-engineering/<repo>",
		},
		{
			"on a one-element module path", "go.mod", "module idea_packer\n",
			"module path idea_packer has no dot in its first element",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			module := fstest.MapFS{"go.mod": goMod, test.file: {Data: []byte(test.content)}}
			_, err := config.New(config.Options{}).For(module, "key")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Errorf("For() error = %v, want one containing %q", err, test.want)
			}
		})
	}
}

func TestForPassesTheImportsKeyToTheLayersLinter(t *testing.T) {
	data, err := config.New(config.Options{}).For(fstest.MapFS{"go.mod": goMod}, "key")
	if err != nil {
		t.Fatal(err)
	}

	var got struct {
		Linters struct {
			Settings struct {
				Custom struct {
					Layers struct {
						Settings map[string]string `yaml:"settings"`
					} `yaml:"layers"`
				} `yaml:"custom"`
			} `yaml:"settings"`
		} `yaml:"linters"`
	}
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if key := got.Linters.Settings.Custom.Layers.Settings["imports"]; key != "key" {
		t.Errorf("layers settings imports = %q, want %q", key, "key")
	}
}

func TestForWithoutGoLintYmlKeepsTheBuiltInSet(t *testing.T) {
	data, err := config.New(config.Options{}).For(fstest.MapFS{"go.mod": goMod}, "key")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "layers") {
		t.Errorf("configuration has no layers linter:\n%s", data)
	}
}
