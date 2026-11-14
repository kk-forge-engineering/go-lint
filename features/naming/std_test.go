package naming

import (
	"go/types"
	"maps"
	"slices"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestStandardMethodsMatchTheStandardLibrary(t *testing.T) {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName | packages.NeedTypes}, "std")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{}
	for _, pkg := range pkgs {
		if pkg.Types == nil || strings.Contains("/"+pkg.PkgPath+"/", "/internal/") || strings.HasPrefix(pkg.PkgPath, "vendor/") {
			continue
		}
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			typeName, isType := scope.Lookup(name).(*types.TypeName)
			if !isType || !typeName.Exported() {
				continue
			}
			iface, isInterface := typeName.Type().Underlying().(*types.Interface)
			if !isInterface {
				continue
			}
			for method := range iface.Methods() {
				if method.Exported() && hasInitialism(method.Name()) {
					signature := signatureOf(method.Signature())
					if !slices.Contains(want[method.Name()], signature) {
						want[method.Name()] = append(want[method.Name()], signature)
					}
				}
			}
		}
	}
	for _, signatures := range want {
		slices.Sort(signatures)
	}
	got := maps.Clone(standardMethods)
	for _, signatures := range got {
		slices.Sort(signatures)
	}
	if !maps.EqualFunc(got, want, slices.Equal) {
		t.Errorf("standardMethods = %#v\nwant %#v", got, want)
	}
}
