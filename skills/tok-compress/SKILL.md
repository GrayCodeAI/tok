---
name: tok-compress
description: >
  Compress natural language memory files (CLAUDE.md, todos, preferences) into tok format
  to save input tokens. Preserves all technical substance, code, URLs, and structure.
  Compressed version overwrites the original file. Human-readable backup saved as FILE.original.md.
  Trigger: /tok:compress <filepath> or "compress memory file"
---

# Tok Compress

## Purpose

Compress natural language files (CLAUDE.md, todos, preferences) into tok-speak to
reduce input tokens. Compressed version overwrites original. Human-readable backup
saved as `<filename>.original.md`.

## Trigger

`/tok:compress <filepath>` or when user asks to compress a memory file.

## Process

The `tok` CLI provides this directly — no separate script or LLM call required.
Prefer the CLI path; it's deterministic and offline.

1. Resolve the absolute path of the target file.

2. Run:

       tok md <absolute_filepath> --mode full

   Other `--mode` values: `lite`, `full` (default), `ultra`, `wenyan-lite`,
   `wenyan-full`, `wenyan-ultra`.

3. The CLI will:
   - preserve code blocks, URLs, file paths, headings, and tables exactly
   - compress prose lines per the selected mode
   - write the compressed version in place
   - back up the original as `<filename>.original.md` (only on first run)
   - print a token-savings summary

4. Report the summary to the user. Mention the backup path.

## Restore

To revert:

    tok md <filepath> --restore

This copies `<filepath>.original.md` back to `<filepath>`.

## Compression Rules

These are encoded in the CLI; this list is informational.

### Remove
- Articles: a, an, the
- Filler: just, really, basically, actually, simply
- Pleasantries: "sure", "certainly", "of course", "happy to"
- Hedging: "it might be worth", "you could consider"
- Redundant phrasing: "in order to" → "to", "due to the fact that" → "because"

### Preserve exactly (never modify)
- Fenced code blocks (``` blocks)
- Inline code (`backtick`)
- URLs and markdown links
- File paths and shell commands
- Headings (keep text; compress body below)
- Tables (full rows preserved)
- Dates, version numbers, numeric literals

### Wenyan modes (CJK-aware)
- Strips modern particles (的, 了, 吧, 呢)
- Maps causal words to arrows (because → →, therefore → →)
- Abbreviates common tech terms (configuration → config, function → fn)

## Boundaries

Refuse to compress:
- Files under `.git/`, `node_modules/`, `vendor/`
- Binary files (non-text)
- Files > 10 MB (warn and ask)

Report errors from the CLI exit code verbatim.
