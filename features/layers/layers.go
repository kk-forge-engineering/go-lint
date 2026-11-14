// Package layers holds a Go module to the layers of the KK Forge architecture standard, as a go/analysis analyzer.
package layers

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis"
)

// analyzer reports each import that breaks the layers or another module's surface, and each cmd/
// package that is more than a launcher. It reports each entities module that fewer than two modules
// of features, orchestration and entities import, and each orchestration module that imports fewer
// than two features. It holds a test file only to the module surfaces and the HTTP framework ban.
var analyzer = &analysis.Analyzer{
	Name: "layers",
	Doc:  "holds a Go module to the layers of the KK Forge architecture standard",
	Run:  run,
}

// allowed lists the layers each layer may import.
var allowed = map[string][]string{
	"entrypoints":   {"entrypoints", "orchestration", "features", "entities", "shared", "infra"},
	"orchestration": {"features", "entities", "shared"},
	"features":      {"entities", "shared"},
	"entities":      {"entities", "shared"},
	"shared":        {"shared"},
	"infra":         {"infra", "shared"},
}

// serverFrameworks lists the modules of the HTTP server frameworks in use in Go. The HTTP controller is the
// entrypoint's adapter, so a framework stays out of every other layer. fasthttp is left out, because infra may use it
// as a client.
var serverFrameworks = []string{
	"github.com/gofiber/fiber",
	"github.com/gin-gonic/gin",
	"github.com/labstack/echo",
	"github.com/go-chi/chi",
	"github.com/gorilla/mux",
	"github.com/julienschmidt/httprouter",
}

// isServerFramework returns whether an import path is a package of an HTTP server framework, at any major version.
func isServerFramework(importPath string) bool {
	return slices.ContainsFunc(serverFrameworks, func(framework string) bool {
		return importPath == framework || strings.HasPrefix(importPath, framework+"/")
	})
}

// published lists the layers whose modules a caller outside reaches only through the module's own package, the
// folder directly under the layer.
var published = []string{"orchestration", "features", "entities"}

func run(pass *analysis.Pass) (any, error) {
	var module goModule
	var packageDir string
	for _, file := range allFiles(pass) {
		name := pass.Fset.Position(file.Package).Filename
		// The loop skips a file outside any Go module. The go tool writes cgo's files and
		// a test binary's main package into its build cache, outside the layers.
		found, err := moduleOf(filepath.Dir(name))
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(found.root, filepath.Dir(name))
		if err != nil {
			return nil, err
		}
		module, packageDir = found, filepath.ToSlash(rel)
		checkFile(pass, module, packageDir, file, strings.HasSuffix(name, "_test.go"))
	}
	return nil, checkModule(pass, module, packageDir)
}

func checkFile(pass *analysis.Pass, module goModule, dir string, file *ast.File, isTest bool) {
	launcher, isLauncher := strings.CutPrefix(dir, "cmd/")
	switch {
	case isTest && (dir == "cmd" || isLauncher):
		checkImports(pass, module, place{layer: "cmd", module: dir}, file, isTest)
	case dir == "cmd":
		pass.Reportf(file.Name.Pos(), "a launcher goes in cmd/<name>/, beside entrypoints/<name>")
	case isLauncher:
		checkLauncher(pass, module, launcher, file)
	default:
		from := placeOf(dir)
		if from == nil {
			return
		}
		checkImports(pass, module, *from, file, isTest)
		// A test may start an httptest server to fake an outside API.
		if !isTest && from.layer != "entrypoints" {
			checkHTTPServer(pass, file)
		}
		// A test drives the clock and the environment itself. A test kit, the package <package>_testing, follows the
		// layer of the package it serves, but it may set them up as a test does.
		isTestKitFile := strings.HasSuffix(file.Name.Name, "_testing")
		if !isTest && !isTestKitFile && from.layer != "infra" && !from.isFeatureInfra() {
			checkEnvironment(pass, *from, file)
		}
		if !isTest && !isTestKitFile && from.layer == "features" && from.dir == from.module && !isGenerated(pass, file) {
			checkSurface(pass, *from, file)
		}
	}
}

// serverHint ends each report of an HTTP server outside entrypoints.
const serverHint = "an HTTP server is an entrypoint's adapter: move it to entrypoints/<name>/http"

// serverNames lists the names of net/http that serve HTTP. A client uses none of them, so infra keeps net/http's
// Request, Client and RoundTripper.
var serverNames = []string{
	"Handler", "HandlerFunc", "ResponseWriter", "ServeMux", "NewServeMux", "Server",
	"ListenAndServe", "ListenAndServeTLS", "Serve", "ServeTLS", "Handle", "HandleFunc",
}

// checkHTTPServer reports each use of a net/http name that serves HTTP, a method such as ServeMux.Handle included. It
// matches the object a name resolves to through the types, a dot-import included. A file the host build leaves out
// has no type information, so this check skips it: run go_lint with that file's GOOS or tags to check it.
func checkHTTPServer(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		ident, isIdent := node.(*ast.Ident)
		if !isIdent {
			return true
		}
		object := pass.TypesInfo.Uses[ident]
		if object != nil && object.Pkg() != nil && object.Pkg().Path() == "net/http" &&
			slices.Contains(serverNames, object.Name()) {
			pass.Reportf(ident.Pos(), "net/http's %s serves HTTP: %s", object.Name(), serverHint)
		}
		return true
	})
}

func checkImports(pass *analysis.Pass, module goModule, from place, file *ast.File, isTest bool) {
	// The composition root, the file compose.go of an entrypoints/<name> package, imports a feature's infra.
	isCompose := from.layer == "entrypoints" && from.dir == from.module && from.dir != "entrypoints" &&
		filepath.Base(pass.Fset.Position(file.Package).Filename) == "compose.go"
	// A scan that fails leaves the list empty, so each entity import of a feature's infra is reported, never passed.
	ports, _ := portsOf(module)
	for _, spec := range file.Imports {
		if problem := module.importProblem(from, spec, isTest, isCompose, ports[from.module]); problem != "" {
			pass.Reportf(spec.Pos(), "%s", problem)
		}
	}
}

// allFiles returns the package's files and the files other systems or build tags would add to it. It leaves out a file
// of another package, such as a generator behind `//go:build ignore`.
func allFiles(pass *analysis.Pass) []*ast.File {
	// The pass of an external test package skips its left-out files, since the package's own
	// pass reads them. A folder whose only external tests stand behind a build tag has no pass
	// of the test package.
	name := pass.Pkg.Name()
	if strings.HasSuffix(name, "_test") {
		return pass.Files
	}
	files := slices.Clone(pass.Files)
	for _, ignored := range pass.IgnoredFiles {
		if !strings.HasSuffix(ignored, ".go") {
			continue
		}
		file, err := parser.ParseFile(pass.Fset, ignored, nil, parser.SkipObjectResolution|parser.ParseComments)
		if err != nil {
			continue
		}
		isExternalTest := strings.HasSuffix(ignored, "_test.go") && file.Name.Name == name+"_test"
		if file.Name.Name == name || isExternalTest {
			files = append(files, file)
		}
	}
	return files
}

type goModule struct {
	root string
	path string
}

// modules caches the Go module that moduleOf found for each folder.
var modules sync.Map

// moduleOf returns the Go module whose go.mod is nearest above dir.
func moduleOf(dir string) (goModule, error) {
	if found, isFound := modules.Load(dir); isFound {
		return found.(goModule), nil
	}
	for root := dir; ; root = filepath.Dir(root) {
		data, err := os.ReadFile(filepath.Join(root, "go.mod"))
		if err == nil {
			module := goModule{root: root, path: modfile.ModulePath(data)}
			modules.Store(dir, module)
			return module, nil
		}
		if filepath.Dir(root) == root {
			return goModule{}, fmt.Errorf("no go.mod above %s", dir)
		}
	}
}

// inside returns an import path's folder from the Go module root, or false when the path is outside the Go module.
func (m goModule) inside(importPath string) (string, bool) {
	if importPath == m.path {
		return ".", true
	}
	return strings.CutPrefix(importPath, m.path+"/")
}

// importProblem returns why an import breaks the layers, or "" when it keeps them. A test file may import any layer,
// such as a real infra adapter in an integration test. It is still held to the module surfaces, and outside
// entrypoints it may not import an HTTP server framework.
func (m goModule) importProblem(from place, spec *ast.ImportSpec, isTest, isCompose bool, portEntities []string) string {
	importPath, _ := strconv.Unquote(spec.Path.Value)
	if from.layer != "entrypoints" && isServerFramework(importPath) {
		return fmt.Sprintf("only entrypoints may import the HTTP server framework %s: %s", importPath, serverHint)
	}
	dir, isInside := m.inside(importPath)
	if !isInside {
		return ""
	}
	to := placeOf(dir)
	if to == nil {
		return ""
	}
	if to.module == from.module {
		if isTest {
			return ""
		}
		return ownModuleProblem(from, *to)
	}
	allowedLayers := allowed[from.layer]
	if from.isFeatureInfra() {
		allowedLayers = featureInfraLayers
	}
	if !isTest && !slices.Contains(allowedLayers, to.layer) {
		if from.isFeatureInfra() {
			return fmt.Sprintf("a feature's infra may not import %s (%s)", to.layer, importPath)
		}
		return fmt.Sprintf("%s may not import %s (%s)", from.layer, to.layer, importPath)
	}
	if !isTest && from.isFeatureInfra() && to.layer == "entities" && !slices.Contains(portEntities, to.module) {
		return fmt.Sprintf("a feature's infra imports only the entities its port %s/port imports (%s)", from.module, importPath)
	}
	// A test also imports a surface's test kit, the package <name>_testing right inside it, as in
	// features/batch/infra/infra_testing for the surface features/batch/infra.
	if isTest && (isTestKit(dir, to.module) || to.isFeatureInfra() && isTestKit(dir, to.module+"/infra")) {
		return ""
	}
	if to.isFeatureInfra() {
		if !isTest && !isCompose {
			return fmt.Sprintf("only the composition root, the compose.go of an entrypoints/<name> package, imports a feature's infra (%s)", importPath)
		}
		if dir != to.module+"/infra" {
			return fmt.Sprintf("import %s/%s/infra, not its inner package %s", m.path, to.module, importPath)
		}
		return ""
	}
	if slices.Contains(published, to.layer) && dir != to.module {
		return fmt.Sprintf("import %s/%s, not its inner package %s", m.path, to.module, importPath)
	}
	return ""
}

// featureInfraLayers lists the layers a feature's infra imports beyond its own feature.
var featureInfraLayers = []string{"entities", "infra", "shared"}

// ownModuleProblem returns why a package may not import another package of its own module, or "". A feature's infra
// imports of its feature only the port, and the feature's other code does not import its infra: the composition root
// passes the adapter in.
func ownModuleProblem(from, to place) string {
	if from.isFeatureInfra() {
		if to.isFeatureInfra() || to.isPort() {
			return ""
		}
		return fmt.Sprintf("a feature's infra imports of its feature only the port %s/port (%s)", to.module, to.dir)
	}
	if to.isFeatureInfra() {
		return fmt.Sprintf("a feature does not import its own infra (%s): the composition root passes the adapter in", to.dir)
	}
	return ""
}

// checkLauncher holds a file in cmd/<launcher>/ to imports of entrypoints and `func main` with one call into
// entrypoints/<launcher>.
func checkLauncher(pass *analysis.Pass, module goModule, launcher string, file *ast.File) {
	entrypoint := ""
	for _, spec := range file.Imports {
		importPath, _ := strconv.Unquote(spec.Path.Value)
		dir, isInside := module.inside(importPath)
		if !isInside || !strings.HasPrefix(dir, "entrypoints/") {
			pass.Reportf(spec.Pos(), "cmd may import only entrypoints (%s)", importPath)
			continue
		}
		if dir == "entrypoints/"+launcher {
			entrypoint = packageName(pass, spec, importPath)
		}
	}
	for _, decl := range file.Decls {
		if gen, isGen := decl.(*ast.GenDecl); isGen && gen.Tok == token.IMPORT {
			continue
		}
		if !isOneCallMain(decl, entrypoint) {
			pass.Reportf(decl.Pos(),
				"cmd/%s may hold only func main with one call, without arguments, into entrypoints/%s", launcher, launcher)
		}
	}
}

// packageName returns the name a file uses for an import: its alias, or the imported package's own name.
func packageName(pass *analysis.Pass, spec *ast.ImportSpec, importPath string) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	for _, imported := range pass.Pkg.Imports() {
		if imported.Path() == importPath {
			return imported.Name()
		}
	}
	return path.Base(importPath)
}

// isOneCallMain returns whether a declaration is `func main() { x.F() }`, where x names the entrypoint's package.
func isOneCallMain(decl ast.Decl, entrypoint string) bool {
	main, isFunc := decl.(*ast.FuncDecl)
	if !isFunc || main.Recv != nil || main.Name.Name != "main" || main.Body == nil || len(main.Body.List) != 1 {
		return false
	}
	statement, isExpr := main.Body.List[0].(*ast.ExprStmt)
	if !isExpr {
		return false
	}
	call, isCall := statement.X.(*ast.CallExpr)
	if !isCall || len(call.Args) != 0 {
		return false
	}
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}
	pkg, isIdent := selector.X.(*ast.Ident)
	return isIdent && entrypoint != "" && pkg.Name == entrypoint
}

// place is where a folder sits in the layers.
type place struct {
	layer string
	// module is the folder directly under the layer, or the layer's own folder for a package there.
	module string
	// dir is the folder itself.
	dir string
}

// inModule returns the path of a place inside its module, such as "infra/db" for features/a/infra/db, or "".
func (p place) inModule() string {
	inside, _ := strings.CutPrefix(p.dir, p.module+"/")
	if inside == p.dir {
		return ""
	}
	return inside
}

// isFeatureInfra returns whether a place sits in a feature's own infra: the package features/<x>/infra or one under it.
func (p place) isFeatureInfra() bool {
	inside := p.inModule()
	return p.layer == "features" && (inside == "infra" || strings.HasPrefix(inside, "infra/"))
}

// isPort returns whether a place is a feature's port: the package features/<x>/port.
func (p place) isPort() bool {
	return p.layer == "features" && p.inModule() == "port"
}

// placeOf returns where a folder from the Go module root sits, or nil outside the layers.
func placeOf(dir string) *place {
	parts := strings.Split(dir, "/")
	layer := parts[0]
	if _, isLayer := allowed[layer]; !isLayer {
		return nil
	}
	if len(parts) == 1 {
		return &place{layer: layer, module: layer, dir: dir}
	}
	return &place{layer: layer, module: layer + "/" + parts[1], dir: dir}
}

// isTestKit returns whether dir is the test kit of the package at surface: the package <name>_testing directly inside
// it, where <name> is the surface's own folder name.
func isTestKit(dir, surface string) bool {
	return dir == surface+"/"+path.Base(surface)+"_testing"
}

// isGenerated returns whether a generator wrote a file, such as a protocol buffer's, which is no code a feature's author
// shapes. cgo's output of an author's file also carries the generated header, but its position maps back to that file,
// so it counts as the author's.
func isGenerated(pass *analysis.Pass, file *ast.File) bool {
	written := pass.Fset.File(file.Package).Name()
	return ast.IsGenerated(file) && written == pass.Fset.Position(file.Package).Filename
}
