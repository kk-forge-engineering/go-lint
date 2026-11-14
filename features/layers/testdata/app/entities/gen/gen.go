// The check puts its report on this file, not on the generated aa_gen.go that sorts first.
// golangci-lint drops reports on generated files, so the report stays visible here.
package gen // want `no features, orchestration or entities module imports entities/gen: move it into the module that uses it, since an entity is what two or more of them share`
