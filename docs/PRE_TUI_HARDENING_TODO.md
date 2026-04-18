# Pre-TUI Hardening TODO

This document tracks the backend cleanup required before the new TokMan TUI.

## Completed

- [x] Remove legacy competitor branding from the repo.
- [x] Replace the old broken TUI and TUI command surface with a clean slate.
- [x] Fix tracking analytics to use the real `commands` schema.
- [x] Add canonical dashboard aggregates in `internal/tracking/dashboard.go`.
- [x] Harden economics, telemetry, and session persistence paths.
- [x] Make agent install/uninstall/status flows use real agent-native integration paths.
- [x] Make `doctor` agent-aware instead of checking only legacy shell-hook assumptions.
- [x] Make `doctor --fix` perform a real repair action by creating default config when missing.
- [x] Add session store summary metrics for diagnostics and future UI composition.
- [x] Replace the `contextread` stub with a real minimal implementation:
  - [x] line slicing
  - [x] signature extraction
  - [x] outline/map mode
  - [x] graph-style summary mode
  - [x] delta snapshots
  - [x] token/line budgeting
- [x] Remove non-core `graph`, `memory`, and `local-llm` command surfaces from the CLI product scope.
- [x] Re-enable skipped Copilot hook rewrite tests.
- [x] Extend `doctor --fix` with safe automated integration repair for broken agent configs.
- [x] Add deeper `doctor` coverage for telemetry store health, integrity baselines, and dashboard data quality.
- [x] Build a canonical composition layer that joins tracking analytics with session-store analytics for the TUI.
- [x] Add richer session analytics:
  - [x] recent session list
  - [x] snapshot history summary
  - [x] active-session context metrics
- [x] Add pricing/catalog completeness checks for provider and model analytics.

## Still Worth Doing Before TUI

- [ ] Add more runtime tests around agent integrations and hook payload variants.

## TUI Build Gate

The new TUI should be built on:

- `internal/tracking/dashboard.go`
- `internal/session` summary/composition APIs
- agent-aware diagnostics from `tokman doctor` / `tokman init --show`
- real integration state, not inferred placeholders
