package layers

import (
	"go/ast"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis"
)

// unitKind tells an app from a package by the folder that holds its Go module, as architecture/core.md lays out a
// monorepo: apps/<name>/ and packages/<name>/, under a root that holds the go.work joining them.
type unitKind string

const (
	appUnit     unitKind = "apps"
	packageUnit unitKind = "packages"
)

// unit is the app or package of a monorepo that a Go module is.
type unit struct {
	kind unitKind
	// repo is the folder that holds apps/ and packages/.
	repo string
}

// unitOf returns the app or package that the Go module at root is, or false when its go.mod sits outside apps/<name>/
// and packages/<name>/ of a folder with a go.work. A single repository cloned under a folder named apps or packages
// has no go.work above it there, so its siblings stay out of its checks.
func unitOf(root string) (unit, bool) {
	kind := unitKind(filepath.Base(filepath.Dir(root)))
	repo := filepath.Dir(filepath.Dir(root))
	if kind != appUnit && kind != packageUnit {
		return unit{}, false
	}
	if _, err := os.Stat(filepath.Join(repo, "go.work")); err != nil {
		return unit{}, false
	}
	return unit{kind: kind, repo: repo}, true
}

// repositoryModules caches, for each repository folder, the kind of each Go module under its apps/ and packages/ by
// the module's path.
var repositoryModules sync.Map

// modulesOf returns the kind of each Go module under a repository's apps/ and packages/, by the path its go.mod
// declares. A folder without a readable go.mod is no Go module and is left out.
func modulesOf(repo string) map[string]unitKind {
	if found, isFound := repositoryModules.Load(repo); isFound {
		return found.(map[string]unitKind)
	}
	kinds := map[string]unitKind{}
	for _, kind := range []unitKind{appUnit, packageUnit} {
		entries, _ := os.ReadDir(filepath.Join(repo, string(kind)))
		for _, entry := range entries {
			data, err := os.ReadFile(filepath.Join(repo, string(kind), entry.Name(), "go.mod"))
			if entry.IsDir() && err == nil {
				kinds[modfile.ModulePath(data)] = kind
			}
		}
	}
	repositoryModules.Store(repo, kinds)
	return kinds
}

// appImported returns the path of the app module an import path sits in, or "" when it sits in none.
func appImported(kinds map[string]unitKind, importPath string) string {
	for modulePath, kind := range kinds {
		if kind == appUnit && (importPath == modulePath || strings.HasPrefix(importPath, modulePath+"/")) {
			return modulePath
		}
	}
	return ""
}

// checkAppImports reports each import of another app's module. Another app's source is forbidden: code two apps use
// moves into a package, and a package imports no app.
func checkAppImports(pass *analysis.Pass, module goModule, file *ast.File) {
	from, isUnit := unitOf(module.root)
	if !isUnit {
		return
	}
	kinds := modulesOf(from.repo)
	for _, spec := range file.Imports {
		importPath, _ := strconv.Unquote(spec.Path.Value)
		app := appImported(kinds, importPath)
		switch {
		case app == "" || app == module.path:
		case from.kind == packageUnit:
			pass.Reportf(spec.Pos(), "a package imports no app (%s)", importPath)
		default:
			pass.Reportf(spec.Pos(), "an import into another app's source is forbidden (%s): "+
				"move the code both apps use into a package", importPath)
		}
	}
}

// isPackageModule returns whether the Go module at root sits under packages/<name>/ of a folder with a go.work.
func isPackageModule(root string) bool {
	found, isUnit := unitOf(root)
	return isUnit && found.kind == packageUnit
}

// checkPackageEnvironment reports each read of the environment in a package's infra. Each app's composition root
// passes the package its config, so the package runs in every app that uses it.
func checkPackageEnvironment(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		ident, isIdent := node.(*ast.Ident)
		if !isIdent {
			return true
		}
		object := packageLevel(pass.TypesInfo.Uses[ident])
		if object == nil {
			return true
		}
		path, name := object.Pkg().Path(), object.Name()
		if path == "os" && environmentNames[path][name] != "" {
			pass.Reportf(ident.Pos(), "%s.%s reads the environment, which a package's infra does not read: "+
				"take the value as a parameter from each app's composition root", path, name)
		}
		return true
	})
}

// repositoryKey returns a line per Go module under the repository's apps/ and packages/ when the Go module at root is
// one of them, so a cache of results changes whenever a module there changes its path.
func repositoryKey(root string) string {
	found, isUnit := unitOf(root)
	if !isUnit {
		return ""
	}
	kinds := modulesOf(found.repo)
	var text strings.Builder
	for _, modulePath := range slices.Sorted(maps.Keys(kinds)) {
		text.WriteString("module " + string(kinds[modulePath]) + " " + modulePath + "\n")
	}
	return text.String()
}
