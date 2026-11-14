# go-lint

- go-lint is its own `main` on golangci-lint's `commands` package. `golangci-lint custom` cannot carry the configuration, and mise cannot install what it builds, so go-lint does not use it.
- `mise run check` must be green before a commit. `mise run lint` runs go-lint from this checkout over its own code.
- `features/config/golangci.yml` is the set every repo runs. A linter joins it only when it catches a bug or a security hole, and its measured findings in our Go repos are real. README.md gives each one's reason. A change to it updates that table.
- A rule change in `layers` or `naming` adds a case with a `// want` comment to its `testdata/app`, and updates its section of README.md.
- `features/naming/std.go` lists the standard library's interface methods with an initialism in capitals. mise installs the newest Go 1.x, so a Go release can add one and turn `TestStandardMethodsMatchTheStandardLibrary` red on any branch: copy in the list it prints.
- README.md → Use lists the flags in `forbidden` in `entrypoints/go_lint/run.go` and the golangci-lint files `features/config/config.go` refuses. A change to either list updates it.
- `entrypoints/go_lint/run_test.go` runs a built go-lint on `testdata/consumer` and expects each issue by file, line and linter, and runs `go_lint vuln` on `testdata/vulnerable`, which needs the network.
- README.md → Release gives the release steps. A change that fails code the last release passed is a new major, with the module path's `/v<major>`.
