# Filter golden-file harness

Each `<case>/` directory is one fixture:

- `cmd.txt` — the shell command the fixture represents
- `input.txt` — raw command output (what the agent would see unfiltered)
- `expected.txt` — what the filter produces today

Running:

```bash
go test ./tests/filters              # verify no regressions
go test ./tests/filters -update      # regenerate expected.txt
```

The harness loads every `*.toml` under `internal/toml/builtin/` and `filters/`,
uses `FindMatchingFilter` to select the rule that `cmd.txt` would hit in
production, runs `TOMLFilterEngine.Apply`, and diffs against `expected.txt`.

## What this harness found

Building this harness uncovered a production bug: filter selection was
non-deterministic because `FindMatchingFilter` and `MatchesCommand` both
iterated Go maps. Two back-to-back `tok git status` calls could pick
different rules and produce different filtered output. Fixed in the same
commit (sort filenames and rule names alphabetically before matching).

## Remaining filter quality gaps

The fixtures pin **current** behavior, not ideal behavior. Known gaps:

| fixture | gap |
|---|---|
| `jest-fail` | drops `Expected "X" but got "Y"` prose (keep rule matches `expect(...)` calls only) |
| `kubectl-get` | passes all pods through — doesn't surface failing ones (CrashLoopBackOff, Error) |

Fixing filters is follow-up work. The harness ensures that when a filter
is improved, the change shows up as a visible `expected.txt` diff that
reviewers can sanity-check — and any future regression blocks CI.

## Adding a fixture

```bash
mkdir tests/filters/my-case
echo "my command args" > tests/filters/my-case/cmd.txt
# paste raw output
$EDITOR tests/filters/my-case/input.txt
go test ./tests/filters -update
# review generated expected.txt — if it drops real signal, that's a filter bug
git add tests/filters/my-case
```
