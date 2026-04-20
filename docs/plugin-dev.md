# Writing tok plugins

A tok plugin is a TOML file that tells tok how to compress the output of a
specific command. There is no Go code to write — a plugin is 10–40 lines of
declarative rules.

## Minimal example

`~/.config/tok/filters/mycmd.toml`:

```toml
schema_version = 1

[mycmd]
match_command = "^mycmd\\b"
strip_ansi = true
strip_lines_matching = [
  "^\\s*$",
  "^Loading config\\.\\.\\.",
]
keep_lines_matching = [
  "^Error:",
  "^PASS",
  "^FAIL",
]
max_lines = 100
```

Drop it into `~/.config/tok/filters/` and any `tok mycmd ...` invocation
(or any shell-hook-captured `mycmd`) will be filtered.

## Schema

| Field | Type | Meaning |
|---|---|---|
| `schema_version` | int | Always `1` today. Required. |
| `[name]` | table | A rule group. Multiple groups may share one file. |
| `match_command` | regex | Regex matched against the full command line. First match wins. |
| `strip_ansi` | bool | Remove ANSI escape codes before other rules. |
| `strip_lines_matching` | [regex] | Drop any line matching any pattern. Applied before `keep_lines_matching`. |
| `keep_lines_matching` | [regex] | If set, only lines matching at least one pattern survive. |
| `replace` | [[table]] | List of `{ pattern = "...", with = "..." }` substitutions (regex). |
| `head` | int | Keep first N lines after filtering. |
| `tail` | int | Keep last N lines after filtering. |
| `max_lines` | int | Hard cap on surviving lines. Overrides `head + tail`. |

## Testing your filter

Each file may include inline tests, parsed and run by `tok doctor`:

```toml
[tests]

[[tests.mycmd]]
name = "drops spinner frames"
input = "Loading config...\nPASS test_alpha\nFAIL test_beta — timeout"
expected = "PASS test_alpha\nFAIL test_beta — timeout"
```

Run:

```bash
tok doctor filters ~/.config/tok/filters/mycmd.toml
```

## Precedence

Filters are resolved in this order, first hit wins:

1. `$TOK_FILTER_DIR` (env var, if set)
2. `~/.config/tok/filters/`
3. `<repo>/filters/`
4. `internal/toml/builtin/` (shipped with the binary)

## Multiple subcommands in one file

```toml
schema_version = 1

[git_log]
match_command = "^git\\s+log\\b"
max_lines = 50

[git_diff]
match_command = "^git\\s+diff\\b"
strip_lines_matching = ["^index [0-9a-f]+"]
```

## Debugging

```bash
TOK_DEBUG=1 tok mycmd --arg      # dump which filter matched
tok compare mycmd --arg           # side-by-side raw vs filtered
```

## Contributing a builtin

To ship a filter with tok for everyone:

1. Put the file in `filters/<name>.toml`.
2. Add at least two inline tests covering the happy path and a failure case.
3. Run `go test ./internal/toml/...` — the builtin loader tests will pick it up.
4. Open a PR with a sample of raw output (≤ 40 lines) and the filtered result.
