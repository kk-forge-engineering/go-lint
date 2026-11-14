// Package naming is a go/analysis analyzer for the KK Forge code style, which writes an initialism in a name as a word:
// HostUrl, not HostURL.
package naming

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports each name a package declares that writes an initialism in capitals, as HostURL does, and each
// import alias that is not camelCase. README.md gives the rules and what the analyzer leaves alone.
var Analyzer = &analysis.Analyzer{
	Name: "naming",
	Doc:  "checks that a name writes an initialism as a word: HostUrl, not HostURL",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	generated := map[*token.File]bool{}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			generated[pass.Fset.File(file.Pos())] = true
		}
	}
	var others map[string][]*types.Func
	for ident, obj := range pass.TypesInfo.Defs {
		if obj == nil || ident.Name == "_" || generated[pass.Fset.File(ident.Pos())] {
			continue
		}
		if _, isImport := obj.(*types.PkgName); isImport {
			checkAlias(pass, ident)
			continue
		}
		if !hasInitialism(ident.Name) {
			continue
		}
		if field, isVar := obj.(*types.Var); isVar && field.Embedded() {
			continue
		}
		if method, isFunc := obj.(*types.Func); isFunc {
			if others == nil {
				others = interfaceMethods(pass.Pkg)
			}
			if implementsOther(method, others) {
				continue
			}
		}
		pass.Reportf(ident.Pos(), "%s: write each initialism as a word, as in %s", ident.Name, asWords(ident.Name))
	}
	return nil, nil
}

// checkAlias reports an import alias that is not camelCase. A package's own name may be snake_case, as in
// convert_request, but an alias is an identifier the file declares.
func checkAlias(pass *analysis.Pass, ident *ast.Ident) {
	name := ident.Name
	if strings.Contains(name, "_") || unicode.IsUpper([]rune(name)[0]) {
		pass.Reportf(ident.Pos(), "%s: write an import alias in camelCase, as in %s", name, camelCase(name))
		return
	}
	if hasInitialism(name) {
		pass.Reportf(ident.Pos(), "%s: write each initialism as a word, as in %s", name, asWords(name))
	}
}

// hasInitialism reports whether a name writes an initialism in capitals.
func hasInitialism(name string) bool {
	return asWords(name) != name
}

// asWords writes each run of capitals as a word: it keeps the first capital and lowers the rest. When a lower-case
// letter follows the run, its last capital starts the next word and stays, so HTTPServer becomes HttpServer, and
// ReadsAValue stays as it is. A lone s after the run is the plural, so URLs becomes Urls, though URLIsValid becomes
// UrlisValid too.
func asWords(name string) string {
	runes := []rune(name)
	for start := 0; start < len(runes); {
		end := start
		for end < len(runes) && unicode.IsUpper(runes[end]) {
			end++
		}
		if end-start < 2 {
			start = max(end, start+1)
			continue
		}
		last := end
		if end < len(runes) && unicode.IsLower(runes[end]) && !isPlural(runes, end) {
			last = end - 1
		}
		for i := start + 1; i < last; i++ {
			runes[i] = unicode.ToLower(runes[i])
		}
		start = end
	}
	return string(runes)
}

// isPlural reports whether the letter at i is an s that ends the word: the name ends there, or no lower-case letter
// follows it.
func isPlural(runes []rune, i int) bool {
	return runes[i] == 's' && (i+1 == len(runes) || !unicode.IsLower(runes[i+1]))
}

// camelCase joins an alias's parts as camelCase words, as in serverUrl for server_url.
func camelCase(name string) string {
	var joined strings.Builder
	for i, part := range strings.FieldsFunc(name, func(r rune) bool { return r == '_' }) {
		runes := []rune(asWords(part))
		if i == 0 {
			runes[0] = unicode.ToLower(runes[0])
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		joined.WriteString(string(runes))
	}
	return joined.String()
}

// interfaceMethods returns the interface methods with an initialism in capitals, keyed by name. It reads the packages
// pkg imports, and those their type information refers to. golangci-lint loads a dependency from export data, which
// refers only to the packages its exported declarations use, so the walk misses a package that a dependency uses only
// inside its function bodies.
func interfaceMethods(pkg *types.Package) map[string][]*types.Func {
	methods := map[string][]*types.Func{}
	seen := map[*types.Package]bool{pkg: true}
	queue := slices.Clone(pkg.Imports())
	for len(queue) > 0 {
		imported := queue[0]
		queue = queue[1:]
		if seen[imported] {
			continue
		}
		seen[imported] = true
		queue = append(queue, imported.Imports()...)
		addInterfaceMethods(methods, imported.Scope())
	}
	return methods
}

// addInterfaceMethods adds to methods each method with an initialism in capitals of an interface the scope declares.
func addInterfaceMethods(methods map[string][]*types.Func, scope *types.Scope) {
	for _, name := range scope.Names() {
		typeName, isType := scope.Lookup(name).(*types.TypeName)
		if !isType {
			continue
		}
		iface, isInterface := typeName.Type().Underlying().(*types.Interface)
		if !isInterface {
			continue
		}
		for method := range iface.Methods() {
			if hasInitialism(method.Name()) {
				methods[method.Name()] = append(methods[method.Name()], method)
			}
		}
	}
}

// implementsOther reports whether a method has the name and the signature of a method of another package's interface:
// one that the package's imports reach, or one of the standard library's. types.Identical ignores the receivers.
func implementsOther(method *types.Func, others map[string][]*types.Func) bool {
	if method.Signature().Recv() == nil {
		return false
	}
	for _, other := range others[method.Name()] {
		if types.Identical(method.Type(), other.Type()) {
			return true
		}
	}
	return slices.Contains(standardMethods[method.Name()], signatureOf(method.Signature()))
}

// signatureOf writes a signature's parameter and result types with their packages' paths, without names or receiver,
// as in func(net/http.ResponseWriter, *net/http.Request).
func signatureOf(signature *types.Signature) string {
	return types.TypeString(types.NewSignatureType(nil, nil, nil, unnamed(signature.Params()),
		unnamed(signature.Results()), signature.Variadic()), nil)
}

func unnamed(tuple *types.Tuple) *types.Tuple {
	vars := make([]*types.Var, tuple.Len())
	for i := range vars {
		vars[i] = types.NewParam(0, nil, "", tuple.At(i).Type())
	}
	return types.NewTuple(vars...)
}
