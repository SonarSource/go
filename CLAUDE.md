# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

A **heavily trimmed fork of [golang/go](https://github.com/golang/go)** (upstream mirror: `go.googlesource.com/go`), maintained at [SonarSource/go](https://github.com/SonarSource/go). It is *not* a full Go distribution — it contains only:

- `src/text/template` (+ `src/text/template/parse`) — the Go text template engine
- `src/fmtsort` — upstream's `internal/fmtsort`, a dependency of `text/template`. It was moved out of `internal/` because packages under an `internal/` path cannot be referenced from another module.

It is published as the Go module `github.com/sonarsource/go/src` and consumed by **sonar-iac**'s `sonar-helm-for-iac` component to parse and evaluate Helm charts. Helm templates *are* Go text templates (plus [Sprig](http://github.com/Masterminds/sprig) helper functions), so sonar-iac reuses this engine rather than reimplementing it.

The pipeline: `sonar-helm-for-iac` (Go binary, stdin/stdout) evaluates Helm templates using this fork, converts the resulting `parse` AST into protobuf (`src/converters/tree_converter.go`), and streams it to the Java analyzer in sonar-iac.

**Everything in this repo outside `src/` is inherited upstream boilerplate** (`README.md`, `LICENSE`, `PATENTS`, `.gitattributes`) and describes the full Go project, not this fork. Don't treat the README as authoritative here.

## Commands

The Go module root is **`src/`, not the repository root**. All Go commands must run from `src/`.

```bash
cd src
go build ./...
go test ./...

# a single package
go test ./text/template/parse/

# a single test, verbose
go test ./text/template/parse/ -run Test_simple_tree -v

go test ./... -coverprofile=coverage.out   # what CI collects
gofmt -l .                                 # must print nothing
```

The Go toolchain version is pinned in `mise.toml` and must match the `go` directive in `src/go.mod`.

## Why the fork exists: the two divergences from upstream

These two changes are the entire reason this fork exists. Preserve them through every upstream merge.

### 1. Comments are in the AST by default

`parse.New()` sets `Mode: ParseComments` (`src/text/template/parse/parse.go`). Upstream defaults to dropping comment nodes; sonar-iac needs them to report issues on commented-out code.

Consequence: upstream's own tests assume comments are skipped, so `parse_test.go` explicitly resets `tree.Mode = 0` before parsing. Expect to apply the same workaround to any upstream test you port in.

### 2. `Pos` is replaced by `Location` (start offset **and** length)

Upstream nodes embed `Pos` (a start offset only) and expose `Position()`. Here, nodes embed a `Location` struct and the `Node` interface exposes:

```go
StartOffset() Pos // byte offset of the node's start
Length()     int  // byte length of the node
```

Upstream only needs positions for error messages; sonar-iac needs precise ranges to highlight issues. `Location`'s fields (`startOffset`, `length`) are **unexported**, so only code inside package `parse` can set them.

**The invariant when touching the parser:** every node constructor takes a `Location` and every node must end up with both a correct start offset *and* a correct length. Lengths are not all set at construction — several are patched up as parsing proceeds:

- `newList` starts with `length: -1`; `ListNode` length is extended in `parse.go` as nodes are appended.
- Pipe, branch (`if`/`range`/`with`) and root lengths are computed after their sub-nodes are known (see the `length =` assignments in `parse.go`).
- `Copy()` methods must carry `Location` across — `ListNode.CopyList` copies `length` explicitly.

If an upstream merge adds a new node type or changes AST shape, you must extend this location logic for it; it will not happen automatically.

### Known location quirks

`src/text/template/parse/locations_test.go` (a Sonar-added test, not upstream) documents the current AST offsets by example and records the accepted imprecisions as `TODO`s — e.g. a comment-only template reports offset 2 rather than 0 because `lexLeftDelim` skips the opening delimiter before `/*`, and branch nodes exclude the `range`/`if` keyword and the trailing `}}`. Read it before "fixing" an offset that looks wrong; it is the clearest spec of intended behaviour.

## Releasing to sonar-iac

Consumers depend on a tagged version, so a change only reaches sonar-iac once it is tagged and pushed.

Because the module path is `github.com/sonarsource/go/src`, Go requires the tag to carry the `src/` subdirectory prefix:

```
src/v<go-version>-<increment>      e.g. src/v1.25.1-1
```

The increment resets per Go version and counts fork revisions on top of it. Then in `sonar-helm-for-iac/go.mod`, bump `github.com/sonarsource/go/src` and run `go mod tidy`.

## Updating to a new upstream Go version

Upstream maintains a `release-branch.go1.xx` branch per release; this fork tracks release branches only (it was originally cut from `release-branch.go1.21`).

1. In an upstream checkout, generate patches limited to the two directories this fork carries:
   ```bash
   git diff go1.23.4..go1.25.1 -- src/text/template      > text-template.patch
   git diff go1.23.4..go1.25.1 -- src/internal/fmtsort   > internal-fmtsort.patch
   ```
   Note the path rename: upstream `src/internal/fmtsort` lives here as `src/fmtsort`.
2. Apply the patches here and resolve conflicts. Conflicts cluster in `parse.go`/`node.go`, because upstream still uses `Pos` where this fork uses `Location`.
3. Re-check the location invariant above for any new or reshaped nodes.
4. Update the Go version in `mise.toml` and `src/go.mod`.
5. Open a PR to `master`, confirm CI is green, then tag and push as described above.

Keep upstream license headers on upstream files unmodified — the fork's license compliance approval depends on it.

## CI

`.github/workflows/build.yml` builds and tests from `src/`, then runs a SonarQube scan against project key `SonarSource_go` on Next, with nightly shadow scans on SonarCloud EU and SonarQube US (also triggerable by adding the `shadow_scan` label to a PR). `iris.yml` runs nightly IRIS analysis across those three platforms. CPD is disabled (`-Dsonar.cpd.exclusions=**`) since the code is largely vendored from upstream.

## Reference

- [Usage of Go fork](https://xtranet-sonarsource.atlassian.net/wiki/spaces/CSD/pages/3564961815/Usage+of+Go+fork) — rationale, license approval, and the canonical upgrade procedure
- [Helm analyzer](https://xtranet-sonarsource.atlassian.net/wiki/spaces/CSD/pages/3036250141/Helm+analyzer) — parent page; see also its *Helm analyzer architecture* and *Evaluating Helm template implementation details* children
- `sonar-helm-for-iac/Readme.md` in sonar-iac-enterprise — the consumer's build and its notes on this dependency
