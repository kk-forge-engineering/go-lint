# go-lint

KK Forge's build of golangci-lint for Go modules, the Go counterpart of `@kk-forge/oxlint-config`. One versioned binary carries the linters, their configuration, the `layers` analyzer and govulncheck, so a repo keeps no `.golangci.yml`, and a laptop and CI check the same rules.

## Use

Add go-lint to the repo's `mise.toml`. Its command is `go_lint`: a name of several words joins its words with `_`. mise builds it with `go install`, in its own module graph, so golangci-lint's dependencies stay out of the repo's `go.mod`:

```toml
[tools]
go = "1"
"go:github.com/kk-forge-engineering/go-lint/v2/cmd/go_lint" = "2"

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
| `infra` | `infra`, `entities`, `shared` |

It reports:

- an import that breaks the table;
- an import of a package inside another `orchestration`, `features` or `entities` module. A module is the folder directly under its layer, and its package there is its surface;
- an import in `cmd/` of anything but a package under `entrypoints/`;
- a package in `cmd/<name>/` that holds more than `func main` with one call, without arguments, into `entrypoints/<name>`;
- a Go file directly in `cmd/`;
- an import of an HTTP server framework (fiber, gin, echo, chi, gorilla/mux, httprouter) outside `entrypoints/`, in a test file too. The HTTP controller is the entrypoint's adapter, at `entrypoints/<name>/http`;
- a use of `net/http`'s server names outside `entrypoints/`: `Handler`, `HandlerFunc`, `ResponseWriter`, `ServeMux`, `NewServeMux`, `Server`, `ListenAndServe`, `ListenAndServeTLS`, `Serve`, `ServeTLS`, `Handle` and `HandleFunc`. A middleware of the form `func(http.Handler) http.Handler` is part of that adapter too. A client keeps `net/http`'s `Request`, `Client` and `RoundTripper`, and a test may start an `httptest` server to fake an outside API.

A test file, in `cmd/` too, is held only to the surface rule and the HTTP server framework rule, so an integration test may import `infra`. The analyzer also reads a package's files for other systems and build tags. It cannot see a folder whose every file the host's build leaves out, such as a package only for Windows: run `GOOS=windows go_lint run ./...` to check one. A package outside the layers and `cmd/` goes unchecked.

## Release

Raise `VERSION` in a PR. After the merge, the owner tags `v<VERSION>` on `main` and pushes the tag.
