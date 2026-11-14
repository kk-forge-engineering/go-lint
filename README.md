# go-lint

KK Forge's build of golangci-lint for Go modules, the Go counterpart of `@kk-forge/oxlint-config`. One versioned binary carries the linters, their configuration, the `layers` and `naming` analyzers and govulncheck, so a repo keeps no `.golangci.yml`, and a laptop and CI check the same rules.

## Use

Add go-lint to the repo's `mise.toml`. Its command is `go_lint`: a name of several words joins its words with `_`. mise builds it with `go install`, in its own module graph, so golangci-lint's dependencies stay out of the repo's `go.mod`:

```toml
[tools]
go = "1"
"go:github.com/kk-forge-engineering/go-lint/v3/cmd/go_lint" = "3"

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
| `naming` | an initialism written in capitals, as in `HostURL`, and an import alias that is not camelCase (below) |
| `errcheck` | an error dropped without a word. Write `_ =` where dropping it is meant |
| `govet` | what every vet analyzer catches. Two stay off: `fieldalignment` reorders fields to save memory, and `shadow` reports the idiomatic shadowed `err` |
| `staticcheck` | its `SA` bug checks. The `S1` simplifications, `ST` style checks and `QF` quick fixes stay off. `ST1003` would ask for `HostURL`, the opposite of `naming` |
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
- an `entities` module that fewer than two modules of `features`, `orchestration` and `entities` import, on the package clause of its first file by name that no generator wrote. A type and its logic live in the one module that uses them, and become an entity only once two modules share them. An entity with one consumer moves into it, or into a feature it uses when the consumer is an `orchestration` module, which holds no logic. An import from another entity counts, since an entity that entities share cannot move into a feature. `shared` imports no entity. An import from `entrypoints` or `infra` does not count, since both use an entity on behalf of a feature: an entrypoint takes the type from that feature's surface, and an infra adapter stays generic behind the feature's port. A test file does not count either, because a test follows its code. Neither does a file of package main, such as a generator behind `//go:build ignore`;
- an `orchestration` module whose files, its inner packages' included, import fewer than two `features` modules, on the package clause of its first file by name that no generator wrote. An orchestration module builds one action from two or more features, and an entrypoint starts a single feature directly. A test file does not count. That an orchestration module holds no logic of its own is for review to hold, since no import shows it;
- a use of `net/http`'s server names outside `entrypoints/`: `Handler`, `HandlerFunc`, `ResponseWriter`, `ServeMux`, `NewServeMux`, `Server`, `ListenAndServe`, `ListenAndServeTLS`, `Serve`, `ServeTLS`, `Handle` and `HandleFunc`. A middleware of the form `func(http.Handler) http.Handler` is part of that adapter too. A client keeps `net/http`'s `Request`, `Client` and `RoundTripper`, and a test may start an `httptest` server to fake an outside API.

A test file, in `cmd/` too, is held only to the surface rule and the HTTP server framework rule, so an integration test may import `infra`. The analyzer also reads a package's files for other systems and build tags. The entities and orchestration checks read the imports of every file under `features/`, `orchestration/` and `entities/`, nested Go modules and `testdata` aside. go_lint passes a key of those imports to the analyzer as a setting, so golangci-lint's cache of results changes whenever a module starts or stops importing an entity or a feature. The analyzer cannot see a folder whose every file the host's build leaves out, such as a package only for Windows: run `GOOS=windows go_lint run ./...` to check one. A package outside the layers and `cmd/` goes unchecked.

## naming

The KK Forge code style writes an initialism as a word, in Go too, against Go's own habit: `HostUrl`, `userId`, `ParseJson`, not `HostURL`, `userID`, `ParseJSON`. naming reports:

- a type, function, method, variable, constant, field, type parameter or label whose name writes an initialism in capitals. The report gives the name written as words. `ReadsAValue` passes: two capitals before a lower-case letter are a one-letter word and the next word. A lone `s` after an initialism is its plural, so `userIDs` gives `userIds`;
- an import alias with an initialism in capitals, or one that is not camelCase, such as `server_url` or `ServerUrl`.

It leaves alone:

- a method with the name and signature of a method of another package's interface, such as `ServeHTTP` of `net/http`'s `Handler`. That package chose the name. naming finds the interfaces of the packages the code imports and of those their exported declarations use, and knows all of the standard library's: `encoding/json` calls `MarshalJSON` on any value that has it, so the method may sit in a package that never imports `encoding/json`. An interface of another library the code does not import, such as a YAML library's `MarshalYAML`, takes `//nolint:naming // implements yaml.Marshaler`;
- an embedded field, which takes its type's name;
- a file marked as generated;
- a struct tag, which is not an identifier. An exported field without a tag gives `encoding/json` its key, so renaming `UserID` to `UserId` changes the key it writes: give the field a tag that keeps the key.

naming cannot tell `serverhttp`, two words joined, from one word, so a code reviewer checks that an import alias of several words is camelCase. Nor can it see the initialism `IP` in `IPv4`: it reads `I` and `Pv4`, so review asks for `Ipv4`. It reads a lone `s` after capitals as a plural, so for `URLIsValid` it suggests `UrlisValid`, and it reports `TreatXAs`, a one-letter word before `As`: write `UrlIsValid`, and give `TreatXAs` `//nolint:naming`.

## Release

Raise `VERSION` in a PR. A release that fails code the last release passed, such as a new rule, is a new major, and a new major changes the module path to `/v<major>` in `go.mod`, every import and the install key above. After the merge, the owner tags `v<VERSION>` on `main` and pushes the tag.
