package layers

import (
	"go/ast"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

// environmentHint ends each report of the environment read outside infra.
const environmentHint = "only infra reads it: take the value as a parameter, or a clock, a random source " +
	"or the config from the composition root"

// environmentImports lists the packages whose import alone reads randomness or writes log output past slog.
var environmentImports = map[string]string{
	"math/rand":    "random numbers",
	"math/rand/v2": "random numbers",
	"crypto/rand":  "random numbers",
	"log":          "log output past slog",
}

// environmentNames lists, per package path, the package-level names that read the clock, its timers or the
// environment. Outside infra no package uses them, an entrypoint included.
var environmentNames = map[string]map[string]string{
	"time": {
		"Now": "the clock", "Since": "the clock", "Until": "the clock", "Sleep": "the clock's timers",
		"After": "the clock's timers", "Tick": "the clock's timers", "NewTimer": "the clock's timers",
		"NewTicker": "the clock's timers", "AfterFunc": "the clock's timers",
	},
	"os": {
		"Getenv": "the environment", "LookupEnv": "the environment", "Environ": "the environment",
		"ExpandEnv": "the environment",
	},
}

// outputNames lists, per package path, the package-level names that write output or build a log handler. Infra and an
// entrypoint use them: an entrypoint prints a CLI's result and sets the default logger.
var outputNames = map[string]map[string]string{
	"fmt": {"Print": "standard output", "Printf": "standard output", "Println": "standard output"},
	"log/slog": {
		"New": "a logger of its own", "NewJSONHandler": "a log handler", "NewTextHandler": "a log handler",
		"DiscardHandler": "a log handler",
	},
}

// checkEnvironment reports each read of the clock, its timers, randomness or the environment, and each write of
// output, in a file outside infra. It matches the package-level object a name resolves to through the types, by its
// package's full path, so a method or a local of the same name passes. A file the host build leaves out has no type
// information, and only its imports are checked here.
func checkEnvironment(pass *analysis.Pass, from place, file *ast.File) {
	for _, spec := range file.Imports {
		importPath, _ := strconv.Unquote(spec.Path.Value)
		if reads, isWatched := environmentImports[importPath]; isWatched {
			pass.Reportf(spec.Pos(), "%s reads or writes %s: %s", importPath, reads, environmentHint)
		}
	}
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
		if reads, isWatched := environmentNames[path][name]; isWatched {
			pass.Reportf(ident.Pos(), "%s.%s reads %s: %s", path, name, reads, environmentHint)
		}
		if writes, isWatched := outputNames[path][name]; isWatched && from.layer != "entrypoints" {
			pass.Reportf(ident.Pos(), "%s.%s writes %s, which only infra and an entrypoint do", path, name, writes)
		}
		return true
	})
}

// packageLevel returns an object when it is declared at its package's top level, such as a function or a variable,
// and nil for a method, a field, a local or a name without type information.
func packageLevel(object types.Object) types.Object {
	if object == nil || object.Pkg() == nil || object.Parent() != object.Pkg().Scope() {
		return nil
	}
	return object
}
