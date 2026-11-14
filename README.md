# go-lint

KK Forge's build of golangci-lint for Go modules, the Go counterpart of `@kk-forge/oxlint-config`. One versioned binary carries the linters, their configuration, the `layers` analyzer and govulncheck, so a repo keeps no `.golangci.yml`, and a laptop and CI check the same rules.

## Use

Add go-lint to the repo's `mise.toml`. Its command is `go_lint`: a name of several words joins its words with `_`. mise builds it with `go install`, in its own module graph, so golangci-lint's dependencies stay out of the repo's `go.mod`:

```toml
[tools]
go = "1"
"go:github.com/kk-forge-engineering/go-lint/v4/cmd/go_lint" = "4"

[tasks.lint]
run = ["go_lint run ./...", "go_lint vuln ./..."]

[tasks."lint:fix"]
run = "go_lint fmt"
```

- `go_lint run [packages]` runs the linters and checks the formatting. It takes golangci-lint's `run` flags. It refuses, in any form, the flags that change the rules or whether the run fails:
  - the set: `-c`, `--config`, `--no-config`, `-E`, `--enable`, `-D`, `--disable`, `--enable-only`, `--default`, `--fast-only`
  - the result: `--issues-exit-code`, `--tests`, `--max-issues-per-linter`, `--max-same-issues`, `--uniq-by-line`
  - the filters to new issues: `-n`, `--new`, `--new-from-rev`, `--new-from-patch`, `--new-from-merge-base`, `--whole-files`
- `go_lint fmt [paths]` formats the code.
- `go_lint vuln [packages]` runs govulncheck: it reports each known vulnerability the code calls, and exits 3 when it finds one. It reads the Go vulnerability database online on each run, so it needs the network. A vulnerable module the code never calls does not fail the run, and neither does a call from a test file. It also checks the standard library of the Go that runs it: a call into a vulnerability that a later Go fixes fails the run until `mise upgrade go` installs that Go. It takes no flag, because a flag could choose another database or let a finding pass.

`go_lint run` and `go_lint fmt` refuse to run beside a `.golangci.yml`, `.yaml`, `.toml` or `.json` at the Go module root.

They also refuse a module whose path has no dot in its first element, such as `idea_packer`. gofumpt takes an import of such a path for the standard library unless it starts like the module's own path, and gci groups it apart, so where one such module imports another, `go_lint run` refuses what `go_lint fmt` writes. Name the module after its repository, such as `github.com/kk-forge-engineering/<repo>`.

## Exclusions

A repo excludes issues in `go_lint.yml` at its Go module root. Each entry is a golangci-lint exclusion rule (`path`, `path-except`, `linters`, `text`, `source`) with a `reason`. go-lint fails when an entry has no reason or the file holds any other key. Paths are read from the Go module root.

```yaml
exclusions:
  - path: tools/
    linters: [errcheck]
    reason: the tools are throwaway scripts a person runs by hand
```

A `//nolint` comment names its linters and gives a reason, as in `//nolint:errcheck // the scratch file may be gone`. go-lint ignores gosec's own `#nosec` comments.

## The set

A linter is in the set when it catches a bug or a security hole and rarely reports anything else, as in `@kk-forge/oxlint-config`. Style is left to the formatters.

| Linter | Catches |
|---|---|
| `layers` | an import against the layer table or another module's surface, and a `cmd/` package that is more than a launcher (below) |
| `errcheck` | an error dropped without a word. Write `_ =` where dropping it is meant |
| `govet` | what every vet analyzer catches. Two stay off: `fieldalignment` reorders fields to save memory, and `shadow` reports the idiomatic shadowed `err` |
| `staticcheck` | its `SA` bug checks. The `S1` simplifications, `ST` style checks and `QF` quick fixes stay off |
| `ineffassign`, `unused` | a value assigned and then overwritten unread, an unused function, type, constant or variable |
| `gocritic` | its diagnostic checks, such as `evalOrder`, `offBy1`, `badRegexp`, `exitAfterDefer` and `deferInLoop`, except `weakCond`, `filepathJoin` and `sloppyReassign`, which mostly report correct code |
| `bodyclose` | an HTTP response body never closed |
| `rowserrcheck` | `rows.Err()` never checked after a `database/sql` loop |
| `errorlint` | an error compared with `==` or asserted to a type, which misses a wrapped error. Its check that `fmt.Errorf` wraps with `%w` stays off: wrapping is a choice about what the error exposes |
| `nilerr`, `nilnesserr` | a `nil` returned where an error was meant |
| `makezero` | `append` to a slice made with a length |
| `asasalint` | a `[]any` passed as one `any` to a variadic function |
| `reassign` | an assignment to another package's variable, such as `io.EOF` |
| `fatcontext` | a context wrapped again on each turn of a loop |
| `nosprintfhostport` | a host and port joined with `Sprintf`, which breaks an IPv6 address |
| `bidichk`, `gocheckcompilerdirectives` | Unicode direction characters that hide code, a `//go:` directive the compiler ignores |
| `testableexamples` | an example that lacks an output comment, so `go test` compiles it but does not run it |
| `gochecknoinits` | an `init` function, which assembles code as a side effect. The KK Forge architecture standard makes assembly an explicit call |
| `noctx` | an HTTP request, a subprocess, a `database/sql` call or a network call made without a context, so a caller cannot cancel it |
| `gosec` `G201`, `G202` | SQL built by joining strings, or by `fmt.Sprintf` into a variable that then runs as a query |
| `gosec` `G401`, `G403`, `G405`–`G407` | a weak hash or cipher, a short RSA key, a fixed IV |
| `gosec` `G402` | TLS that skips certificate checks, allows an old version or uses a weak cipher suite |
| `gosec` `G108` | a profiling endpoint that `net/http/pprof` exposes |
| `gosec` `G112`, `G114` | an HTTP server that sets no timeouts |
| `gosec` `G124` | a cookie that leaves out `Secure`, `HttpOnly` or `SameSite` |
| `nolintlint` | a `//nolint` that leaves out its linter or its reason |

`noctx` and `gosec` skip test files: a test runs under its own deadline and serves no traffic. gosec's other rules stay off, because in our code they mostly report correct code.

The formatters are `gofumpt`, a stricter `gofmt`, and `gci`, which groups imports into the standard library, other modules and the module's own packages.

## layers

The direction guard of the KK Forge architecture standard. Each layer imports only these layers:

| Layer | Imports |
|---|---|
| `entrypoints` | every layer |
| `orchestration` | `features`, `entities`, `shared` |
| `features` | `entities`, `shared` |
| `entities` | `entities`, `shared` |
| `shared` | `shared` |
| `infra` | `infra`, `shared` |

It reports:

- an import that breaks the table;
- in a feature's own infra, `features/<x>/infra` and the packages under it, an import of its feature other than the port package `features/<x>/port`, of an entity the port does not import, or of a layer other than `entities`, `infra` and `shared`. The feature's other packages may not import that infra. Of the other layers only the composition root, the `compose.go` of an `entrypoints/<name>` package, may, through the package `features/<x>/infra` itself. A test file may import it too, and a surface's test kit, the package `<name>_testing` right inside it, such as `features/<x>/<x>_testing` or `features/<x>/infra/infra_testing`;
- outside `infra` and a feature's infra, `entrypoints` included, a read of the clock, its timers, randomness or the environment: `time.Now`, `time.Since`, `time.Until`, `time.Sleep`, `time.After`, `time.Tick`, `time.NewTimer`, `time.NewTicker` and `time.AfterFunc`; `os.Getenv`, `os.LookupEnv`, `os.Environ` and `os.ExpandEnv`; an import of `math/rand`, `math/rand/v2`, `crypto/rand` or `log`, whose output bypasses `slog`. Outside `infra` and `entrypoints`, `fmt.Print`, `fmt.Printf` and `fmt.Println`, `slog.New` and `slog`'s handlers: an entrypoint prints a CLI's result and sets the default logger. A feature may call `slog`'s package-level functions. The check matches the package-level object a name resolves to through the types, by its package's full path, so a method such as `time.Time.After` or a local of the same name passes. Logic takes the value as a parameter, and the composition root hands a clock, a random source or the config to the code that needs one. A test file is exempt, and so is a test kit, the package `<package>_testing`, which otherwise follows the layer of the package it serves;
- an import of a package inside another `orchestration`, `features` or `entities` module. A module is the folder directly under its layer, and its package there is its surface;
- on a feature's surface, the package `features/<x>`, an exported function or variable other than its use cases: `type UseCases` and `func New(o Options) *UseCases`, both in `<x>_use_cases.go`, which the check names from the feature's folder exactly. A scenario is a method of `UseCases`, and a helper stays unexported. A type, a method and a constant pass, and so do a test, a test kit and a file a generator wrote. A feature that only shows has no use cases file;
- an import in `cmd/` of anything but a package under `entrypoints/`;
- a package in `cmd/<name>/` that holds more than `func main` with one call, without arguments, into `entrypoints/<name>`;
- a Go file directly in `cmd/`;
- an import of an HTTP server framework (fiber, gin, echo, chi, gorilla/mux, httprouter) outside `entrypoints/`, in a test file too. The HTTP controller is the entrypoint's adapter, at `entrypoints/<name>/http`;
- an `entities` module that fewer than two modules of `features`, `orchestration` and `entities` import, on the package clause of its first file by name that no generator wrote. A type and its logic live in the one module that uses them, and become an entity only once two modules share them. An entity with one consumer moves into it, or into a feature it uses when the consumer is an `orchestration` module, which holds no logic. An import from another entity counts, since an entity that entities share cannot move into a feature. `shared` imports no entity. An import from `entrypoints` does not count, since an entrypoint takes the type from the feature's surface. An import from a feature's infra counts as that feature, and the root `infra` imports no entity. A test file does not count either, because a test follows its code. Neither does a file of package main, such as a generator behind `//go:build ignore`;
- an `orchestration` module whose files, its inner packages' included, import fewer than two `features` modules, on the package clause of its first file by name that no generator wrote. An orchestration module builds one action from two or more features, and an entrypoint starts a single feature directly. A test file does not count. That an orchestration module holds no logic of its own is for review to hold, since no import shows it;
- a use of `net/http`'s server names outside `entrypoints/`: `Handler`, `HandlerFunc`, `ResponseWriter`, `ServeMux`, `NewServeMux`, `Server`, `ListenAndServe`, `ListenAndServeTLS`, `Serve`, `ServeTLS`, `Handle` and `HandleFunc`. A middleware of the form `func(http.Handler) http.Handler` is part of that adapter too. A client keeps `net/http`'s `Request`, `Client` and `RoundTripper`, and a test may start an `httptest` server to fake an outside API.

A test file, in `cmd/` too, is held only to the surface rule and the HTTP server framework rule, so an integration test may import `infra`. The analyzer also reads a package's files for other systems and build tags, for their imports. Such a file has no type information, so the server and environment checks, which read names through the types, skip it: run `GOOS=windows go_lint run ./...` to check one. The entities and orchestration checks read the imports of every file under `features/`, `orchestration/` and `entities/`, nested Go modules and `testdata` aside. go_lint passes a key of those imports, and of each feature's port package, to the analyzer as a setting, so golangci-lint's cache of results changes whenever a module starts or stops importing an entity or a feature. The analyzer cannot see a folder whose every file the host's build leaves out, such as a package only for Windows, so the same run checks it. A package outside the layers and `cmd/` goes unchecked.

## Release

Raise `VERSION` in a PR. A release that fails code the last release passed, such as a new rule, is a new major, and a new major changes the module path to `/v<major>` in `go.mod`, every import and the install key above. After the merge, the owner tags `v<VERSION>` on `main` and pushes the tag.
