package layers

import (
	"go/ast"
	"go/token"
	"path"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
)

// checkSurface reports what the package features/<x> exports beyond its use cases and its types. The use cases are
// `type UseCases` and `func New(o Options) *UseCases`, both in <x>_use_cases.go. Any other exported function or
// variable fails. A method and a constant pass, since each belongs to a type.
func checkSurface(pass *analysis.Pass, from place, file *ast.File) {
	useCasesFile := path.Base(from.module) + "_use_cases.go"
	in := surfaceFile{
		pass:           pass,
		module:         from.module,
		useCasesFile:   useCasesFile,
		isUseCasesFile: filepath.Base(pass.Fset.Position(file.Package).Filename) == useCasesFile,
	}
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			in.checkFunc(decl)
		case *ast.GenDecl:
			in.checkNames(decl)
		}
	}
}

// surfaceFile is one file of the package features/<x> that checkSurface reads.
type surfaceFile struct {
	pass *analysis.Pass
	// module is the feature's folder, such as features/texts.
	module string
	// useCasesFile is the name of the file that holds the use cases, such as texts_use_cases.go.
	useCasesFile   string
	isUseCasesFile bool
}

func (in surfaceFile) checkFunc(decl *ast.FuncDecl) {
	pass, useCasesFile := in.pass, in.useCasesFile
	if decl.Recv != nil || !decl.Name.IsExported() {
		return
	}
	switch {
	case decl.Name.Name != "New":
		pass.Reportf(decl.Name.Pos(), "%s is a free function on the surface of %s: make it a method of UseCases in %s, "+
			"or unexport it as a helper", decl.Name.Name, in.module, useCasesFile)
	case !in.isUseCasesFile:
		pass.Reportf(decl.Name.Pos(), "New goes in %s, beside UseCases", useCasesFile)
	case !isUseCasesConstructor(decl.Type):
		pass.Reportf(decl.Name.Pos(), "New takes the options and returns the use cases: func New(o Options) *UseCases")
	}
}

func (in surfaceFile) checkNames(decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		switch spec := spec.(type) {
		case *ast.ValueSpec:
			if decl.Tok != token.VAR {
				continue
			}
			for _, name := range spec.Names {
				if name.IsExported() {
					in.pass.Reportf(name.Pos(), "%s is a variable on the surface of %s, which gives out only UseCases and "+
						"types: unexport it, or make it a method of UseCases", name.Name, in.module)
				}
			}
		case *ast.TypeSpec:
			if spec.Name.Name == "UseCases" && !in.isUseCasesFile {
				in.pass.Reportf(spec.Name.Pos(), "UseCases goes in %s, the file named for its feature", in.useCasesFile)
			}
		}
	}
}

// isUseCasesConstructor returns whether a function type is `func(o Options) *UseCases`, with the parameter named or not.
func isUseCasesConstructor(function *ast.FuncType) bool {
	if function.TypeParams != nil || function.Results == nil {
		return false
	}
	params, results := function.Params.List, function.Results.List
	if len(params) != 1 || len(params[0].Names) > 1 || len(results) != 1 || len(results[0].Names) > 0 {
		return false
	}
	param, isIdent := params[0].Type.(*ast.Ident)
	result, isPointer := results[0].Type.(*ast.StarExpr)
	if !isIdent || !isPointer {
		return false
	}
	useCases, isIdent := result.X.(*ast.Ident)
	return param.Name == "Options" && isIdent && useCases.Name == "UseCases"
}
