---
name: update-upstream-go
description: >
  Merge a new upstream Go release into this fork of golang/go, which carries only
  text/template and fmtsort. Replays upstream's diff on top of the fork's two deliberate
  divergences (comments kept in the AST, Pos replaced by Location) and verifies they
  survived. Trigger on phrases like "update the fork to Go 1.28," "merge upstream
  go1.27.2," "bump the forked Go version," or "update to a new upstream Go version."
---

# Updating the fork to a new upstream Go version

This fork carries only `src/text/template` (+ `parse`) and `src/fmtsort`. An upgrade means
replaying upstream's diff for those directories on top of the fork's two deliberate
divergences, then proving the divergences survived.

Read `CLAUDE.md` at the repo root first — it is the spec for *why* the divergences exist.
This skill is the *how*.

## The two divergences you are protecting

1. **Comments are in the AST by default.** `parse.New()` sets `Mode: ParseComments`.
   Upstream drops comment nodes, so upstream tests that assume comments vanish will fail
   here.
2. **`Pos` → `Location`** (start offset *and* length). Every node constructor takes a
   `Location`; `Node` exposes `StartOffset()`/`Length()`. Conflicts cluster here.

## Procedure

### 1. Get an upstream checkout

A blobless clone is enough and takes well under a minute:

```bash
git clone --filter=blob:none --no-checkout https://go.googlesource.com/go /tmp/go-upstream
git -C /tmp/go-upstream tag -l 'go1.28*'   # confirm the target tag exists
```

Determine the current version from `mise.toml` (it must match the `go` directive in
`src/go.mod`). That is the `FROM` tag.

### 2. Generate the patches

```bash
cd /tmp/go-upstream
git diff goFROM..goTO -- src/text/template      > /tmp/text-template.patch
git diff goFROM..goTO -- src/internal/fmtsort   > /tmp/internal-fmtsort.patch
git diff --stat goFROM..goTO -- src/text/template src/internal/fmtsort
```

Note the path rename: upstream `src/internal/fmtsort` is `src/fmtsort` here. `fmtsort`
rarely changes — if that patch is empty, skip it. If it is not, strip the `internal/`
path component before applying.

`src/text/template/*` paths match the fork layout exactly, so those apply at the repo
root with `-p1` and no rewriting.

### 3. Apply per file, never as one blob

**Do not run `git apply --reject` on the combined patch.** It aborts partway and leaves
the working tree with files *deleted* rather than writing `.rej` files. (Recover with
`git checkout -- src`.)

Split the patch per file and triage:

```bash
cd /tmp/go-upstream
for f in <files from the --stat above>; do
  git diff goFROM..goTO -- "src/text/template/$f" > "/tmp/p-$(echo $f | tr / -).patch"
done

cd <repo root>
for p in /tmp/p-*.patch; do
  printf '%-45s ' "$(basename $p)"
  git apply --check "$p" 2>/dev/null && echo CLEAN || echo CONFLICT
done
```

Apply every CLEAN patch with `git apply`. For each CONFLICT, use `patch(1)`, which
*does* write usable rejects:

```bash
patch -p1 --no-backup-if-mismatch < /tmp/p-parse-node.go.patch
cat src/text/template/parse/node.go.rej   # then hand-apply, then delete the .rej
```

Most rejects are false alarms: the hunk's *context* lines mention `Pos` where the fork
says `Location`, while the changed lines themselves are divergence-neutral. Apply the
upstream change verbatim and keep the fork's `Location` context.

### 4. Re-check the location invariant

Only needed if upstream added or reshaped a node type. Confirm mechanically rather than
by eye:

```bash
git -C /tmp/go-upstream show goTO:src/text/template/parse/node.go \
  | grep -E '^\tNode[A-Za-z]+|^type [A-Za-z]+Node|^func \(t \*Tree\) new[A-Za-z]+' | sort > /tmp/up-nodes.txt
grep -E '^\tNode[A-Za-z]+|^type [A-Za-z]+Node|^func \(t \*Tree\) new[A-Za-z]+' \
  src/text/template/parse/node.go | sort > /tmp/fork-nodes.txt
diff /tmp/up-nodes.txt /tmp/fork-nodes.txt
```

The expected output is *only* `pos Pos` vs `loc Location` constructor signatures. Any new
`NodeXxx` constant, `XxxNode` type or `newXxx` constructor means you must extend the
location logic for it — a correct start offset **and** a correct length, including any
patch-up in `parse.go` and in `Copy()`. See the `length =` assignments in `parse.go`.

### 5. Bump the version

```bash
sed -i '' 's/^go = "FROM"$/go = "TO"/' mise.toml
sed -i '' 's/^go FROM$/go TO/' src/go.mod
```

These two must stay in sync. CI installs Go via `jdx/mise-action` reading `mise.toml`, so
`.github/workflows/` needs no edit — but re-grep for stray version strings
(`grep -rn 'FROM' --exclude-dir=.git .`) in case that ever changes.

### 6. Build and test

```bash
cd src
go build ./... && go test ./...
gofmt -l .        # must print nothing
go vet ./...
```

### 7. Fix upstream tests that assume comments are dropped

This is the predictable failure. Two remedies, in order of preference:

- **Adapt the expectation to include the comment** when the test is really about
  round-tripping or AST shape. This is usually *better* coverage than upstream's, because
  it exercises comment-node paths upstream cannot reach. Say so in a comment.
- **Reset `tree.Mode = 0`** before parsing when the test is about something else entirely
  and the comment is incidental. See `parse_test.go: testParse` for the established
  pattern. Only works where the test drives `parse.Tree` directly — tests going through
  `template.New(...).Parse(...)` have no handle on `Mode`.

Always leave a comment naming the divergence, so the next upgrade does not "fix" it back.

### 8. Prove nothing was missed, and nothing broke downstream

Diff every fork file against upstream, normalising the fork's import paths:

```bash
rm -rf /tmp/up && mkdir -p /tmp/up
git -C /tmp/go-upstream archive goTO src/text/template | tar -x -C /tmp/up
cd /tmp/up/src/text/template
find . -name '*.go' | while read f; do
  fork="<repo>/src/text/template/$f"
  [ -f "$fork" ] || { echo "MISSING IN FORK: $f"; continue; }
  sed -e 's#"github.com/sonarsource/go/src/text/template/parse"#"text/template/parse"#' \
      -e 's#"github.com/sonarsource/go/src/text/template"#"text/template"#' \
      -e 's#"github.com/sonarsource/go/src/fmtsort"#"internal/fmtsort"#' "$fork" > /tmp/norm.go
  diff -q "$f" /tmp/norm.go >/dev/null || echo "DIFFERS: $f"
done
```

Every reported difference must be explainable as an intended divergence. Confirm anything
you did not write this session is pre-existing with
`git show HEAD:<path> | grep ...` rather than assuming. Known pre-existing ones:

- `link_test.go` is absent from the fork (dropped in `ba24ffbfea`).
- `exec_test.go` expects error column `t:1:2`, not upstream's `t:1:23` — a consequence of
  the `Location` change.

Then check the consumer is unaffected, by diffing the exported API across the upgrade:

```bash
git stash -q
go doc -all ./text/template/parse > /tmp/api-before.txt; go doc -all ./text/template > /tmp/api-before-tpl.txt
git stash pop -q
go doc -all ./text/template/parse > /tmp/api-after.txt;  go doc -all ./text/template > /tmp/api-after-tpl.txt
diff /tmp/api-before.txt /tmp/api-after.txt; diff /tmp/api-before-tpl.txt /tmp/api-after-tpl.txt
```

Doc-comment-only differences are fine. A signature change means sonar-iac's
`sonar-helm-for-iac/src/converters/tree_converter.go` may need updating — check which
`parse.*` symbols it uses before concluding.

### 9. Commit

Branch and commit subject both use the Jira key; search Jira for the ticket rather than
inventing one (`project in (SONARIAC, SONARGO) AND summary ~ "Go"`). Past upgrades:
SONARIAC-1885 (1.23.4), SONARIAC-2242 (1.25), SONARIAC-2793 (1.27.1).

```
<KEY> Update forked Go to <version>
```

The body should record what upstream changed, which hunks conflicted and why, and any
test adapted for a divergence — that is the context the *next* upgrade needs.

Do not push, tag, or open the PR without asking.

## Follow-up after the PR merges

The change only reaches sonar-iac once tagged. Because the module path is
`github.com/sonarsource/go/src`, the tag carries the subdirectory prefix:

```
src/v<go-version>-<increment>      e.g. src/v1.27.1-1
```

The increment resets per Go version. Then bump `github.com/sonarsource/go/src` in
`sonar-helm-for-iac/go.mod` and run `go mod tidy`. The Jira ticket usually covers both
repos, so it should not be closed on the fork PR alone.
