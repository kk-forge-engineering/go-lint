package layers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// scannedLayers lists the layers whose imports the module-wide checks count. An entrypoint or an infra adapter that
// imports an entity uses it on behalf of a feature, and a test follows its code, so none of them shares an entity.
var scannedLayers = []string{"features", "orchestration", "entities"}

// moduleImports maps each module of scannedLayers, by its folder from the Go module root, to the other modules of
// features and entities its files import, sorted.
type moduleImports map[string][]string

// scans caches a *scan for each Go module root, so a run reads each Go module once.
var scans sync.Map

type scan struct {
	once    sync.Once
	imports moduleImports
	err     error
}

// importsOf returns the module imports of the Go module. It reads every file in features, orchestration and entities,
// for every system and build tag, but no test file and no file of package main, such as a generator behind
// `//go:build ignore`. It skips a nested Go module and the files and folders the go tool ignores.
func importsOf(module goModule) (moduleImports, error) {
	entry, _ := scans.LoadOrStore(module.root, &scan{})
	cached := entry.(*scan)
	cached.once.Do(func() { cached.imports, cached.err = scanImports(module) })
	return cached.imports, cached.err
}

func scanImports(module goModule) (moduleImports, error) {
	found := moduleImports{}
	fset := token.NewFileSet()
	for _, layer := range scannedLayers {
		layerRoot := filepath.Join(module.root, layer)
		err := filepath.WalkDir(layerRoot, func(path string, entry fs.DirEntry, err error) error {
			if errors.Is(err, fs.ErrNotExist) && path == layerRoot {
				return filepath.SkipDir
			}
			if err != nil {
				return err
			}
			name := entry.Name()
			if entry.IsDir() {
				if isSkippedDir(path, name) {
					return filepath.SkipDir
				}
				return nil
			}
			// The scan skips a file that is not a regular file, since a FIFO or a symlink to a device would stall the read.
			if !entry.Type().IsRegular() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") ||
				isIgnoredName(name) {
				return nil
			}
			return scanFile(module, fset, path, found)
		})
		if err != nil {
			return nil, err
		}
	}
	for _, modules := range found {
		slices.Sort(modules)
	}
	return found, nil
}

// scanFile adds the features and entities modules that one file imports to found. It skips a file of package main,
// and a file that does not parse, which the compiler reports.
func scanFile(module goModule, fset *token.FileSet, path string, found moduleImports) error {
	rel, err := filepath.Rel(module.root, filepath.Dir(path))
	if err != nil {
		return err
	}
	from := placeOf(filepath.ToSlash(rel))
	if from == nil {
		return nil
	}
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil || file.Name.Name == "main" {
		return nil //nolint:nilerr // The compiler reports a file that does not parse.
	}
	for _, spec := range file.Imports {
		importPath, _ := strconv.Unquote(spec.Path.Value)
		dir, isInside := module.inside(importPath)
		if !isInside {
			continue
		}
		to := placeOf(dir)
		if to == nil || (to.layer != "entities" && to.layer != "features") || to.module == from.module ||
			slices.Contains(found[from.module], to.module) {
			continue
		}
		found[from.module] = append(found[from.module], to.module)
	}
	return nil
}

// isIgnoredName returns whether the go tool ignores a file or folder by its name.
func isIgnoredName(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// isSkippedDir returns whether the scan skips a folder: one the go tool ignores, or the root of a nested Go module.
func isSkippedDir(path, name string) bool {
	if name == "testdata" || name == "vendor" || isIgnoredName(name) {
		return true
	}
	_, err := os.Stat(filepath.Join(path, "go.mod"))
	return err == nil
}

// consumers returns the modules that import the module `to`, sorted.
func (imports moduleImports) consumers(to string) []string {
	var found []string
	for _, from := range slices.Sorted(maps.Keys(imports)) {
		if slices.Contains(imports[from], to) {
			found = append(found, from)
		}
	}
	return found
}

// features returns the features modules that the module `from` imports, sorted.
func (imports moduleImports) features(from string) []string {
	return slices.DeleteFunc(slices.Clone(imports[from]), func(module string) bool {
		return !strings.HasPrefix(module, "features/")
	})
}

// checkModule reports an entities module that fewer than two modules of features, orchestration and entities import,
// and an orchestration module that imports fewer than two features. It reports once per module, on the package clause
// of reportFile in the module's own package.
func checkModule(pass *analysis.Pass, module goModule, dir string) error {
	at := placeOf(dir)
	if at == nil || at.module != dir || strings.HasSuffix(pass.Pkg.Name(), "_test") ||
		(at.layer != "entities" && at.layer != "orchestration") {
		return nil
	}
	first := reportFile(pass)
	if first == nil {
		return nil
	}
	imports, err := importsOf(module)
	if err != nil {
		return err
	}
	if at.layer == "orchestration" {
		reportOrchestration(pass, first, dir, imports.features(dir))
	} else {
		reportEntity(pass, first, dir, imports.consumers(dir))
	}
	return nil
}

// reportFile returns the file a module report lands on: the first non-test file by name, for any system and build tag,
// that no generator wrote. golangci-lint drops an issue in a generated file, so a generated file takes the report only
// when the package has no other.
func reportFile(pass *analysis.Pass) *ast.File {
	var first *ast.File
	isBefore := func(file *ast.File) bool {
		if first == nil {
			return true
		}
		if generated, firstGenerated := ast.IsGenerated(file), ast.IsGenerated(first); generated != firstGenerated {
			return !generated
		}
		return pass.Fset.Position(file.Package).Filename < pass.Fset.Position(first.Package).Filename
	}
	for _, file := range allFiles(pass) {
		if !strings.HasSuffix(pass.Fset.Position(file.Package).Filename, "_test.go") && isBefore(file) {
			first = file
		}
	}
	return first
}

// entityHint ends each report of an entities module with fewer than two consumers.
const entityHint = "since an entity is what two or more of them share"

func reportEntity(pass *analysis.Pass, file *ast.File, entity string, consumers []string) {
	switch len(consumers) {
	case 0:
		pass.Reportf(file.Name.Pos(),
			"no features, orchestration or entities module imports %s: move it into the module that uses it, %s", entity, entityHint)
	case 1:
		// An orchestration module holds no logic of its own, so its entity moves into a feature it uses.
		destination := consumers[0]
		if strings.HasPrefix(destination, "orchestration/") {
			destination = "a feature that " + destination + " uses"
		}
		pass.Reportf(file.Name.Pos(), "only %s imports %s: move it into %s, %s", consumers[0], entity, destination, entityHint)
	}
}

// orchestrationHint ends each report of an orchestration module with fewer than two features.
const orchestrationHint = "since an orchestration module builds one action from two or more features"

func reportOrchestration(pass *analysis.Pass, file *ast.File, orchestration string, features []string) {
	switch len(features) {
	case 0:
		pass.Reportf(file.Name.Pos(), "%s imports no feature: move it into the feature or entrypoint that uses it, %s",
			orchestration, orchestrationHint)
	case 1:
		pass.Reportf(file.Name.Pos(),
			"%s imports only %s: move it into %s, and let the entrypoint start that feature directly, %s",
			orchestration, features[0], features[0], orchestrationHint)
	}
}

// ImportsKey returns a key of the imports the module-wide checks of the Go module at root count. It changes whenever a
// module starts or stops importing a feature or an entity, so a cache of results can key on it.
func ImportsKey(root string) (string, error) {
	module, err := moduleOf(root)
	if err != nil {
		return "", err
	}
	imports, err := importsOf(module)
	if err != nil {
		return "", err
	}
	var text strings.Builder
	for _, from := range slices.Sorted(maps.Keys(imports)) {
		text.WriteString(from + ":" + strings.Join(imports[from], ",") + "\n")
	}
	sum := sha256.Sum256([]byte(text.String()))
	return hex.EncodeToString(sum[:]), nil
}
