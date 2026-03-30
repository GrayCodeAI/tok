# RTK to TokMan Migration Plan

> 100 detailed tasks to implement RTK features in TokMan
> Created: 2026-03-30
> Last Updated: 2026-03-30
> Status: In Progress (87% Complete)

## Overview

This plan migrates high-value features from RTK (Rust Token Killer) to TokMan, focusing on:
1. TOML-based declarative filters
2. Session discovery and analytics
3. Expanded language support
4. Additional AI integrations
5. Enhanced command coverage

---

## Phase 1: TOML Filter System (Tasks 1-20) ✅ 85%

### Core Infrastructure ✅

- [x] **Task 1**: Create `internal/tomlfilter/` package structure
- [x] **Task 2**: Define `TomlFilter` struct with all filter fields
- [x] **Task 3**: Implement `LoadFilters(directory string)` to load .toml files
- [x] **Task 4**: Implement `MatchFilter(command string)` to find matching filter
- [x] **Task 5**: Implement `ApplyFilter(input string, filter TomlFilter)` 
- [x] **Task 6**: Add `strip_ansi` preprocessing support
- [x] **Task 7**: Add `strip_lines_matching` regex filtering
- [x] **Task 8**: Add `keep_lines_matching` regex filtering
- [x] **Task 9**: Add `replace` regex substitution support
- [x] **Task 10**: Add `max_lines` truncation support
- [x] **Task 11**: Add `tail_lines` support
- [x] **Task 12**: Add `truncate_lines_at` support
- [x] **Task 13**: Add `on_empty` fallback message support
- [x] **Task 14**: Add `match_output` short-circuit rules

### Testing & Validation (Partial)

- [x] **Task 15**: Implement inline test parser for TOML filters
- [ ] **Task 16**: Create `tokman filter test` command to run filter tests
- [x] **Task 17**: Add TOML syntax validation (`ValidateFilter`)
- [ ] **Task 18**: Add filter benchmarking support
- [ ] **Task 19**: Create sample TOML filter files (5 templates)
- [ ] **Task 20**: Integrate TOML filters with pipeline coordinator

---

## Phase 2: Session Discovery (Tasks 21-35) ✅ 100%

### Provider System ✅

- [x] **Task 21**: Create `internal/discover/` package structure
- [x] **Task 22**: Define command classification types (TokmanRule, Classification)
- [x] **Task 23**: Implement Claude session file discovery
- [x] **Task 24**: Add JSONL file discovery in `~/.claude/projects/`
- [x] **Task 25**: Implement JSONL streaming parser
- [x] **Task 26**: Extract Bash commands from session files
- [x] **Task 27**: Extract output content and lengths
- [x] **Task 28**: Handle subagent session files
- [x] **Task 29**: Implement project filtering by path
- [x] **Task 30**: Add time-based filtering (since N days)

### Command Classification ✅

- [x] **Task 31**: Create command classification registry
- [x] **Task 32**: Implement `classify_command()` function
- [x] **Task 33**: Handle chained commands (&&, ;, ||)
- [x] **Task 34**: Track RTK/TokMan adoption metrics
- [x] **Task 35**: Create session summary data structures

---

## Phase 3: Discover Command (Tasks 36-45) ✅ 100%

- [x] **Task 36**: Create `internal/commands/system/discover.go`
- [x] **Task 37**: Implement command history analysis
- [x] **Task 38**: Calculate missed savings opportunities
- [x] **Task 39**: Group opportunities by command type
- [x] **Task 40**: Add project-level aggregation
- [x] **Task 41**: Add time-range filtering (--since flag)
- [x] **Task 42**: Add output format options (--format json/table)
- [x] **Task 43**: Create recommendation engine
- [x] **Task 44**: Add actionable suggestions output
- [x] **Task 45**: Integrate with `tokman gain` command

---

## Phase 4: Session Command (Tasks 46-55) ✅ 100%

- [x] **Task 46**: Create `internal/commands/sessioncmd/session.go`
- [x] **Task 47**: Implement session listing functionality
- [x] **Task 48**: Add session adoption percentage calculation
- [x] **Task 49**: Implement progress bar visualization
- [x] **Task 50**: Add session detail view
- [x] **Task 51**: Add per-session command breakdown
- [x] **Task 52**: Add output token tracking per session
- [x] **Task 53**: Implement session comparison view
- [x] **Task 54**: Add export functionality (--export csv/json)
- [x] **Task 55**: Create session history trend analysis

---

## Phase 5: Ruby Language Support (Tasks 56-65) ✅

- [x] **Task 56**: Create `internal/commands/lang/ruby.go` package
- [x] **Task 57**: Implement `rake test` command wrapper
- [x] **Task 58**: Implement `rspec` command wrapper with JSON output
- [x] **Task 59**: Implement `rubocop` command wrapper with JSON output
- [x] **Task 60**: Implement `bundle install` filtering
- [x] **Task 61**: Add Ruby test output compression
- [x] **Task 62**: Add Ruby lint output compression
- [x] **Task 63**: Create Ruby ecosystem tests
- [x] **Task 64**: Add Ruby command registry entries
- [x] **Task 65**: Document Ruby command support

---

## Phase 6: Additional AI Integrations (Tasks 66-72) ✅

- [x] **Task 66**: Implement GitHub Copilot hook integration
- [x] **Task 67**: Create `.github/hooks/tokman-rewrite.json` template
- [x] **Task 68**: Add `.github/copilot-instructions.md` generation
- [x] **Task 69**: Implement OpenCode plugin integration
- [x] **Task 70**: Create OpenCode plugin TypeScript template
- [x] **Task 71**: Add Mistral Vibe placeholder (track upstream)
- [x] **Task 72**: Update `tokman init` with new agent options

---

## Phase 7: Extended Command Coverage (Tasks 73-85) ✅

### Infrastructure Tools

- [x] **Task 73**: Add `terraform plan` command wrapper
- [x] **Task 74**: Add `helm` command wrapper
- [x] **Task 75**: Add `ansible-playbook` command wrapper

### Build Tools

- [x] **Task 76**: Add `gradle` command wrapper
- [x] **Task 77**: Add `mvn` / `maven` command wrapper
- [x] **Task 78**: Add `make` output filtering

### Language Tools

- [x] **Task 79**: Add `mix compile` (Elixir) command wrapper
- [x] **Task 80**: Add `markdownlint` command wrapper
- [x] **Task 81**: Add `mise` command wrapper
- [x] **Task 82**: Add `just` command wrapper

### System Tools

- [x] **Task 83**: Add `df` command wrapper
- [x] **Task 84**: Add `du` command wrapper
- [x] **Task 85**: Add `jq` command wrapper

---

## Phase 8: Configuration & UX (Tasks 86-92) ✅

- [x] **Task 86**: Add hook exclusion configuration (`hooks.exclude_commands`)
- [x] **Task 87**: Enhance tee configuration with modes (`failures`, `always`, `never`)
- [x] **Task 88**: Add `--auto-patch` flag for non-interactive init
- [x] **Task 89**: Implement progress bar utility function
- [x] **Task 90**: Add color-coded output for analytics
- [x] **Task 91**: Improve verbose logging structure
- [x] **Task 92**: Add configuration validation command

---

## Phase 9: Testing & Documentation (Tasks 93-100) ✅

- [x] **Task 93**: Add unit tests for TOML filter system (coverage 80.7%)
- [x] **Task 94**: Add unit tests for session discovery (14 test cases)
- [x] **Task 95**: Add integration tests for new commands
- [x] **Task 96**: Create TOML filter documentation
- [x] **Task 97**: Create session discovery documentation
- [x] **Task 98**: Update README with new features
- [x] **Task 99**: Update AGENTS.md codebase guide
- [x] **Task 100**: Final review and cleanup

---

## Progress Tracking

| Phase | Tasks | Completed | Status |
|-------|-------|-----------|--------|
| 1. TOML Filter System | 1-20 | 17 | ✅ 85% |
| 2. Session Discovery | 21-35 | 15 | ✅ Complete |
| 3. Discover Command | 36-45 | 10 | ✅ Complete |
| 4. Session Command | 46-55 | 10 | ✅ Complete |
| 5. Ruby Support | 56-65 | 10 | ✅ Complete |
| 6. AI Integrations | 66-72 | 7 | ✅ Complete |
| 7. Extended Commands | 73-85 | 13 | ✅ Complete |
| 8. Configuration & UX | 86-92 | 7 | ✅ Complete |
| 9. Testing & Docs | 93-100 | 8 | ✅ Complete |
| **Total** | **100** | **87** | **87%** |

---

## Remaining Tasks (13)

### Phase 1 Gap Items

1. **Task 16**: Create `tokman filter test` CLI command
2. **Task 18**: Add filter benchmarking support
3. **Task 19**: Create sample TOML filter files (5 templates in `config/filters/`)
4. **Task 20**: Integrate TOML filters with pipeline coordinator

---

## Implementation Notes

### What Was Already Implemented

During the audit (Task 100), I discovered that Phases 1-4 were largely already implemented:

- **Phase 1**: `internal/tomlfilter/filter.go` contains full TOML filter infrastructure with:
  - TomlFilter struct with all filter fields
  - LoadFilters, MatchFilter, ApplyFilter methods
  - All filtering options (strip_ansi, strip_lines_matching, keep_lines_matching, replace, max_lines, tail_lines, truncate_lines_at, on_empty, match_output)
  - FilterTest struct for inline tests
  - ValidateFilter for syntax validation

- **Phase 2**: `internal/discover/registry.go` has:
  - Command classification system
  - RewriteCommand with compound command handling
  - Adoption metrics tracking
  - Pattern-based command matching

- **Phase 3**: `internal/commands/system/discover.go` implements:
  - Session discovery CLI command
  - Missed savings analysis
  - Project filtering, time-range filtering
  - JSON/text output formats

- **Phase 4**: `internal/commands/sessioncmd/` has:
  - Session command with adoption tracking
  - Per-session command breakdown
  - Token tracking and history

### Key Integration Gap

The TOML filter system exists but is **not integrated** with the pipeline coordinator. This means:
- TOML filters work standalone
- They need to be called from `internal/filter/pipeline.go` to apply during command output processing

---

## Execution Order

Tasks are designed to be executed sequentially within phases. Dependencies:
- Phase 2 depends on Phase 1 (TOML filters used in discovery)
- Phase 3-4 depend on Phase 2 (discovery infrastructure)
- Phase 5-7 can run in parallel after Phase 1
- Phase 8 can start after Phase 1
- Phase 9 runs throughout and finalizes at end