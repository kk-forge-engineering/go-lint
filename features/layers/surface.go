package layers

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

// checkSurface reports what the package features/<x> exports beyond its use cases, its types and its sentinel errors.
// The use cases are `type UseCases` and `func New(o Options) *UseCases`, both in <x>_use_cases.go, which holds the two
// whenever it exists. Any other exported function or variable fails. A method and a constant pass, since each belongs
// to a type.
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
	if in.isUseCasesFile && (!in.hasUseCases || !in.hasNew) {
		pass.Reportf(file.Name.Pos(), "%s holds type UseCases and func New(o Options) *UseCases", useCasesFile)
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
	// hasUseCases and hasNew record whether the use cases file declares type UseCases and func New.
	hasUseCases, hasNew bool
}

func (in *surfaceFile) checkFunc(decl *ast.FuncDecl) {
	pass, useCasesFile := in.pass, in.useCasesFile
	if decl.Recv != nil || !decl.Name.IsExported() {
		return
	}
	in.hasNew = in.hasNew || (decl.Name.Name == "New" && in.isUseCasesFile)
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

func (in *surfaceFile) checkNames(decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		switch spec := spec.(type) {
		case *ast.ValueSpec:
			if decl.Tok != token.VAR {
				continue
			}
			for i, name := range spec.Names {
				// A file the host build leaves out has no types, so no check can tell a sentinel there. A run with its
				// GOOS or build tags checks it, as it does for the server and environment checks.
				if !name.IsExported() || in.pass.TypesInfo.Defs[name] == nil || in.isSentinelError(spec, i) {
					continue
				}
				in.pass.Reportf(name.Pos(), "%s is a variable on the surface of %s, which gives out only UseCases, types "+
					"and sentinel errors: unexport it, or make it a method of UseCases", name.Name, in.module)
			}
		case *ast.TypeSpec:
			if spec.Name.Name == "UseCases" && in.isUseCasesFile {
				in.hasUseCases = true
			}
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

// isSentinelError returns whether the i-th name of a variable declaration is a sentinel error, which a caller matches
// with errors.Is: a variable of type error set by errors.New, or by fmt.Errorf with a constant format that wraps no
// error. It reads the types.
func (in *surfaceFile) isSentinelError(spec *ast.ValueSpec, i int) bool {
	variable, isVar := in.pass.TypesInfo.Defs[spec.Names[i]].(*types.Var)
	if !isVar || !types.Identical(variable.Type(), types.Universe.Lookup("error").Type()) ||
		len(spec.Values) != len(spec.Names) {
		return false
	}
	call, isCall := ast.Unparen(spec.Values[i]).(*ast.CallExpr)
	if !isCall {
		return false
	}
	function := typeutil.StaticCallee(in.pass.TypesInfo, call)
	if function == nil || function.Pkg() == nil {
		return false
	}
	switch function.Pkg().Path() + "." + function.Name() {
	case "errors.New":
		return true
	case "fmt.Errorf":
		if len(call.Args) == 0 {
			return false
		}
		format := in.pass.TypesInfo.Types[call.Args[0]].Value
		return format != nil && format.Kind() == constant.String && !wraps(constant.StringVal(format))
	}
	return false
}

// wraps returns whether a format of fmt.Errorf wraps an error: whether one of its verbs is %w, with any flags, width,
// precision or argument index, as in %[1]w. A %% prints a percent sign and is no verb.
func wraps(format string) bool {
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		i++
		for i < len(format) && strings.IndexByte("+-# 0123456789.*[]", format[i]) >= 0 {
			i++
		}
		if i < len(format) && format[i] == 'w' {
			return true
		}
	}
	return false
}
