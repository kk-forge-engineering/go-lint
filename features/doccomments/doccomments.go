// Package doccomments holds Go doc comments to the `//` form of the KK Forge comments standard, as a go/analysis
// analyzer.
package doccomments

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// analyzer reports a `/* */` comment in the doc comment of a package clause, a declaration, a spec or a field. godoc
// and editors read either form, so the compiler and the other linters let a block pass.
var analyzer = &analysis.Analyzer{
	Name: "doccomments",
	Doc:  "reports a /* */ block in a doc comment, which the KK Forge comments standard writes as // lines",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.File:
				report(pass, node.Doc)
			case *ast.GenDecl:
				// The check skips an import declaration and leaves its `/* */` block in place. cgo reads the C code of a package
				// from the block above `import "C"`.
				if node.Tok == token.IMPORT {
					return false
				}
				report(pass, node.Doc)
			case *ast.FuncDecl:
				report(pass, node.Doc)
			case *ast.TypeSpec:
				report(pass, node.Doc)
			case *ast.ValueSpec:
				report(pass, node.Doc)
			case *ast.Field:
				report(pass, node.Doc)
			}
			return true
		})
	}
	return nil, nil
}

// report reports each `/* */` comment of a doc comment.
func report(pass *analysis.Pass, doc *ast.CommentGroup) {
	if doc == nil {
		return
	}
	for _, comment := range doc.List {
		if strings.HasPrefix(comment.Text, "/*") {
			pass.Reportf(comment.Pos(), "a doc comment is // lines, not a /* */ block")
		}
	}
}
