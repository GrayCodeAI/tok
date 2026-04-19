# Tok World-Class TUI Rebuild Plan

## Goal

Rebuild Tok's terminal UI from scratch as a premium token intelligence product using the Charm stack:

- `Bubble Tea v2`
- `Bubbles v2`
- `Lip Gloss v2`
- `Huh`
- `Glamour`
- optional `Harmonica`

The new TUI should feel like a polished product, not a generic metrics screen and not a classic sysadmin console.

Implementation source of truth:

- [docs/TUI_PRODUCT_SPEC.md](/Users/lakshmanpatel/Desktop/ProjectAlpha/tok/docs/TUI_PRODUCT_SPEC.md:1)

## Codebase-First Findings

This plan is based on inspection of the current Tok codebase, especially:

- `internal/tracking`
- `internal/session`
- `internal/commands/core`
- `internal/commands/hooks`
- `internal/commands/system`
- `internal/mcp`
- `internal/config`
- `docs/AGENT_INTEGRATION.md`

### What Tok Already Has

The codebase already supports more TUI-relevant analytics than a fresh design would normally assume:

- SQLite tracking database at `tracking.db`
- main `commands` table with:
  - `command`
  - `original_tokens`
  - `filtered_tokens`
  - `saved_tokens`
  - `project_path`
  - `session_id`
  - `exec_time_ms`
  - `timestamp`
  - `parse_success`
  - `agent_name`
  - `model_name`
  - `provider`
  - `model_family`
  - `context_kind`
  - `context_mode`
  - `context_resolved_mode`
  - `context_target`
  - `context_related_files`
  - `context_bundle`
- indexes for agent, provider, model, and context fields
- `agent_summary` SQL view
- per-layer stats table: `layer_stats`
- cost aggregation table: `cost_aggregations`
- filter metrics table: `filter_metrics`
- parse failure table
- checkpoint events table
- session system with its own `sessions.db`
- `SessionManager` with session metadata and context blocks
- `tok gain` command for savings analytics
- `tok proxy` for pass-through execution with tracking
- `tok mcp` for MCP-based usage
- `tok init` with multi-agent integration setup
- native hook processors under `tok hook`
- Claude Code usage correlation via `tok ccusage` and `tok cc-economics`

### Existing Integration Model in Code

Tok already models itself as a layer between:

- coding agents
- hooks / shell execution
- MCP clients
- provider/model traffic

That is visible in:

- env-based attribution in `internal/tracking/tracker.go`
- agent integration commands in `internal/commands/core/init.go`
- hook processors in `internal/commands/hooks/hook.go`
- MCP server in `internal/mcp`
- session model in `internal/session`

### Important Current Gaps

The rebuild must account for some inconsistencies and unfinished areas:

1. Some analytics code still queries `command_history`, but the live schema uses `commands`.
2. Session data lives in a separate `sessions.db`, so session analytics require cross-source composition.
3. Cost tracking exists, but model pricing coverage is incomplete and should be normalized before premium dashboards rely on it.
4. Integration attribution is strong for agent/provider/model, but IDE/editor attribution is not clearly persisted yet.
5. Existing CLI analytics are functional in some places, but not yet shaped for dashboard-grade, high-confidence aggregates.

### Implication

The TUI rebuild should not begin with rendering work.

It must begin with:

- data correctness
- aggregate definition
- attribution normalization
- cross-database composition for sessions

## Product Positioning

Tok TUI should become:

- the observability layer between coding agents/IDEs and LLM providers
- a token savings dashboard
- a provider/model cost intelligence tool
- a compression analytics cockpit
- a developer productivity and gamification layer
- a detailed session and command drilldown interface

Tok's real role is:

- coding agent or IDE issues work
- Tok sits in the middle
- Tok observes, compresses, routes, and measures
- the upstream system is a coding agent CLI, agent TUI, plugin, or IDE
- the downstream system is an LLM provider/model

That means the product must understand and attribute activity across:

- agent
- IDE/editor
- provider
- model
- project
- session
- command
- pipeline

It should answer two classes of questions:

1. Executive summary:
   - How much token volume did Tok save?
   - How much money did it save?
   - Which provider, model, command, project, and session drove the savings?
   - Am I trending better or worse than yesterday/week/month?

2. Deep operational analysis:
   - Which layers are doing useful work?
   - Which commands are still wasteful?
   - Which provider/model combinations are expensive?
   - Which coding agents and IDEs are producing the most waste or the most savings?
   - Which agents, sessions, and projects are most efficient?
   - Where are the biggest opportunities for additional reduction?

## Integration Model

Tok should support attribution from any of these sources:

- coding agent CLI
- coding agent TUI
- IDE plugin
- editor integration
- shell hook
- proxy mode
- API proxy mode

Examples of upstream clients:

- Codex
- Claude Code
- Cursor
- Cline
- Copilot
- Gemini CLI or IDE integrations
- Windsurf
- generic MCP-connected tooling

Examples of downstream providers:

- OpenAI
- Anthropic
- Google
- xAI
- OpenRouter
- Ollama
- LM Studio
- provider-compatible self-hosted endpoints

## Attribution Dimensions

Every tracked unit should attempt to capture:

- agent name
- agent type
- IDE/editor name
- integration channel
- provider name
- model name
- model family
- project path
- session id
- command or operation type
- timestamp
- original token count
- filtered token count
- saved token count
- estimated cost before
- estimated cost after
- estimated savings

## Required Product Capability

The new TUI must be able to answer:

- Which coding agent saved the most tokens?
- Which IDE integration generated the highest token volume?
- Which provider cost the most?
- Which model was most expensive before Tok filtering?
- Which model/provider pair benefited most from Tok?
- Which agent/provider combinations are inefficient?
- Which projects are producing the highest AI spend?
- Which sessions from a given coding agent were best or worst?

## Design Principles

### 1. Dashboard First

The home screen must feel like a real product dashboard:

- strong hero metrics
- clear trend lines
- compact but legible charts
- token and cost intelligence up front
- minimal visual noise

### 2. Progressive Disclosure

The first screen should be easy to understand in seconds.
Deep details should exist, but be one step deeper.

### 3. No Fake Filler

No giant empty cards, decorative boxes, or placeholder chrome.
Every visible region must carry information or action value.

### 4. Terminal Native

The TUI should feel fast, keyboard-first, and robust under resize.
It should degrade cleanly on smaller terminals.

### 5. Product-Level Visual Identity

The UI should feel intentional:

- one clear visual language
- consistent spacing and typography hierarchy
- restrained motion
- strong data contrast
- premium rather than noisy

## Information Architecture

## Global Navigation

Primary sections:

1. `Home`
2. `Easy Day`
3. `Analytics`
4. `Providers`
5. `Models`
6. `Agents`
7. `Integrations`
8. `Sessions`
9. `Commands`
10. `Pipeline`
11. `Rewards`
12. `Logs`
13. `Config`

Secondary overlays:

- global search
- help
- command/session detail
- provider/model detail
- chart fullscreen
- export modal

## Screen Definitions

## 1. Home

Purpose:
The main token intelligence dashboard.

Must show:

- total tokens before
- total tokens after
- total tokens saved
- total reduction percentage
- total estimated cost before
- total estimated cost after
- total cost saved
- today saved
- 7-day saved
- 30-day saved
- current streak
- token saver score
- best provider today
- best command today
- biggest win today
- top current alert/opportunity

Visual structure:

- hero KPI row
- cost + token trend charts
- provider comparison block
- top commands block
- streak and rewards block
- recent wins feed

## 2. Easy Day

Purpose:
5-second digest mode.

Must answer:

- how much did I save today?
- how much money did I save today?
- am I on a streak?
- which provider/model was best today?
- which command saved the most today?
- am I above or below normal efficiency?
- what is the single biggest next opportunity?

Visual structure:

- summary cards
- one trend sparkline
- one recommendation card
- one motivation/gamification card

## 3. Analytics

Purpose:
Deep time-series analysis.

Must include:

- tokens before vs after over time
- reduction % over time
- cost saved over time
- session efficiency trend
- command efficiency trend
- compression layer contribution over time
- cache contribution over time
- project-level savings trend

Controls:

- date range
- grouping by hour/day/week/month
- filter by provider/model/agent/project

## 4. Providers

Purpose:
Provider-level savings and cost analysis.

Must show:

- total usage by provider
- total saved by provider
- cost before/after by provider
- cost saved by provider
- reduction rate by provider
- sessions by provider
- command mix by provider

Provider detail should also show:

- top upstream agents using that provider
- top IDE integrations using that provider
- top models under that provider
- biggest spenders before reduction
- biggest savers after reduction

Providers include:

- OpenAI
- Anthropic
- Google
- xAI
- Ollama
- LM Studio
- other detected providers

## 5. Models

Purpose:
Model-level efficiency analysis.

Must show:

- model name
- provider
- commands count
- original token volume
- filtered token volume
- saved token volume
- reduction %
- estimated cost before
- estimated cost after
- estimated cost saved
- average session efficiency
- top agents using the model
- top IDEs using the model

Examples:

- `gpt-4.1`
- `gpt-5`
- `claude-sonnet`
- `claude-opus`
- `gemini`
- local models

## 6. Agents

Purpose:
Agent-level attribution and comparison.

Must show:

- agent name
- total commands
- total sessions
- total original tokens
- total filtered tokens
- total saved tokens
- reduction %
- estimated cost before
- estimated cost after
- cost saved
- top provider used by this agent
- top model used by this agent
- top project used by this agent
- streak or efficiency score for this agent

Examples:

- Codex
- Claude Code
- Cursor
- Cline
- Copilot
- Gemini
- custom integrations

## 7. Integrations

Purpose:
Track where Tok is being used from.

Must show:

- integration channel
- shell/CLI usage
- IDE/editor usage
- plugin usage
- proxy usage
- MCP usage when applicable
- token volume by integration
- savings by integration
- cost saved by integration

Examples:

- terminal CLI
- VS Code
- Cursor IDE
- JetBrains
- shell hook
- API proxy

## 8. Sessions

Purpose:
Session drilldown.

Must show:

- session id
- start/end time
- duration
- agent name
- IDE/editor or integration source
- provider
- model
- commands count
- reads count
- original tokens
- filtered tokens
- saved tokens
- reduction %
- estimated cost saved
- quality or anomaly hints

Session detail view:

- timeline
- top commands
- layer impact
- savings by step
- errors/warnings
- upstream tool attribution
- downstream provider/model attribution

## 9. Commands

Purpose:
Command-level savings intelligence.

Must show:

- command
- frequency
- average original tokens
- average filtered tokens
- average saved tokens
- reduction %
- total saved tokens
- estimated cost saved
- provider/model mix
- agent mix
- IDE/integration mix
- last seen

Command detail:

- historical trend
- raw vs filtered sample
- layer contribution
- command-specific opportunities

## 10. Pipeline

Purpose:
Compression pipeline observability.

Must show:

- per-layer enablement
- layer effectiveness
- total tokens saved by layer
- average contribution
- last used
- layer interaction notes
- skip/early-exit stats

Detail:

- entropy layer impact
- perplexity layer impact
- compaction impact
- H2O impact
- cache impact
- failures or quality risks

## 11. Rewards

Purpose:
Gamification and motivation.

Must show:

- daily streak
- weekly streak
- monthly streak
- lifetime token saver points
- current level
- next milestone
- badges
- achievements
- best day
- best session
- best command win
- leaderboard hooks for future team mode

## 12. Logs

Purpose:
Operational live diagnostics.

Must include:

- searchable log stream
- level filters
- source filters
- live/frozen mode
- selected-entry detail
- copy/export support

## 13. Config

Purpose:
Explain the current system state.

Must show:

- effective config values
- pipeline preset
- enabled layers
- budget settings
- tracking DB path
- cache settings
- provider pricing assumptions
- dashboard mode settings
- version/build info

## Metrics Model

## Token Metrics

Core token fields:

- original tokens
- filtered tokens
- saved tokens
- reduction percentage
- average reduction
- median reduction
- percentile reductions

## Cost Metrics

Per provider/model:

- input price
- output price if relevant
- effective estimated cost before
- effective estimated cost after
- cost saved
- cost saved per command
- cost saved per session
- cost saved per day/week/month

Per agent/integration:

- cost before
- cost after
- cost saved
- token volume
- savings ratio
- model/provider usage mix

## Efficiency Metrics

- commands/hour
- tokens saved/hour
- cost saved/hour
- average tokens saved per command
- average reduction percentage by provider/model
- efficiency rank by command
- cache-assisted savings ratio

## Quality Metrics

If available now or later:

- parse success rate
- fallback rate
- compaction usage rate
- safe output pass rate
- replay anomaly count

## Required Data Sources

## Existing Likely Sources

- `internal/tracking`
- `internal/config`
- `internal/economics`
- `internal/cache`
- `internal/session`
- `internal/filter`
- telemetry/logging packages

Existing attribution hints already likely available or partially present:

- `AgentName`
- `ModelName`
- `Provider`
- `ModelFamily`

Concrete existing surfaces already confirmed in code:

- `commands` table in `internal/tracking/migrations.go`
- `agent_summary` SQL view in `internal/tracking/migrations.go`
- session metadata in `internal/session/types.go`
- session database schema in `internal/session/manager.go`
- provider/model/agent attribution capture in `internal/tracking/tracker.go`
- proxy tracking in `internal/commands/core/proxy.go`
- MCP server in `internal/commands/core/mcp.go` and `internal/mcp`
- Claude Code economics bridge in `internal/commands/core/economics.go`
- Claude Code usage import via `internal/commands/system/ccusage.go`

## Data That May Need New Collection

- IDE/editor attribution
- integration channel attribution
- per-layer contribution metrics per command
- provider/model cost tables normalized for analytics
- streak snapshot state
- rewards/achievement state
- daily aggregation tables
- model/provider attribution consistency
- session timeline events

## Data Correctness Phase

Before any new Bubble Tea screens are implemented, fix and validate analytics semantics:

1. Replace stale `command_history` queries with `commands`-backed queries in the analytics layer.
2. Audit all savings and cost commands against the current schema.
3. Confirm whether `cost_aggregations` and `filter_metrics` are actually populated in runtime flows.
4. Define a canonical attribution model for:
   - agent
   - integration
   - provider
   - model
   - project
   - session
5. Decide whether session summaries should remain cross-database or be denormalized into tracking aggregates.
6. Add tests around dashboard-critical aggregate queries.

## Recommended New Data Model Additions

New analytical aggregates:

- `daily_token_aggregates`
- `provider_usage_aggregates`
- `model_usage_aggregates`
- `agent_usage_aggregates`
- `integration_usage_aggregates`
- `command_efficiency_aggregates`
- `layer_effectiveness_aggregates`
- `session_summaries`
- `reward_state`
- `achievement_events`

## Gamification Model

## Points

Proposed base scoring:

- +1 point per 1,000 tokens saved
- multiplier for high reduction %
- bonus for daily streak continuation
- bonus for “best day” record
- bonus for “provider optimization” improvements

## Streaks

Types:

- daily savings streak
- weekly goal streak
- monthly efficiency streak

Definitions:

- daily streak continues if saved tokens exceed a minimum threshold
- weekly streak continues if weekly goal is met
- efficiency streak continues if reduction % stays above threshold

## Achievements

Examples:

- first 10K saved
- first 100K saved
- first 1M saved
- 7-day streak
- 30-day streak
- 90-day streak
- provider optimizer
- command master
- pipeline champion
- session sprinter

## Screen Layout Strategy

## Home Layout

- top hero: 6 KPI cards
- middle left: token trend chart
- middle right: cost trend chart
- lower left: provider/model leaderboard
- lower center: agent/integration leaderboard
- lower right: streaks and badges
- footer: live activity feed

## Detail Layout

- left rail for scope/filter
- main content with charts/table
- right inspector pane for selected row
- bottom shortcut bar

## Visual Style Direction

- black or near-black base with disciplined accent usage
- premium cyan/green/amber/red semantic palette
- restrained borders
- emphasis on spacing, typography hierarchy, and signal density
- motion only where meaningful:
  - KPI refresh
  - streak/achievement reveal
  - loading transitions

## Technical Architecture

## UI Layer

- Bubble Tea model tree
- screen router
- shared layout primitives
- reusable KPI card components
- reusable chart widgets
- reusable leaderboard blocks
- reusable detail panes

## State Layer

- app state store
- view state
- filter state
- selection state
- periodic refresh commands

## Data Layer

- tracker readers
- aggregation service
- cost calculation service
- reward engine
- trend service
- provider/model normalization service

## Testing Strategy

- Bubble Tea snapshot/golden tests
- small-screen render tests
- navigation state tests
- analytics computation tests
- streak logic tests
- cost calculation tests

Charm testing reference:

- use Bubble Tea testing patterns and golden tests
- keep layout deterministic by fixing terminal dimensions during tests

## Delivery Phases

## Phase 0: Data Correctness and Foundation

- fix stale analytics queries still using `command_history`
- validate existing aggregate-producing commands
- normalize provider/model pricing assumptions
- confirm current agent attribution quality
- define integration attribution schema additions
- define session/tracking composition strategy

Success criteria:

- every dashboard metric can be traced to a correct query
- provider/model/agent attribution is trustworthy
- no premium UI depends on unverified placeholder math

## Phase 1: UI Foundation

- remove old TUI
- choose architecture
- define data contracts
- define theme system
- define screen routing

Success criteria:

- Bubble Tea app shell exists
- design tokens and layout primitives are stable
- loading, resize, and error states are designed first

## Phase 2: Dashboard V1

- Home
- Easy Day
- Analytics basics
- Providers basics
- Models basics
- Agents basics
- Integrations basics
- Logs basics

Success criteria:

- token and cost visibility works
- trends render correctly
- no broken layouts

## Phase 3: Drilldowns

- Sessions
- Commands
- Pipeline
- fullscreen detail overlays
- search/filter/export

Success criteria:

- user can explain any major saving with drilldown

## Phase 4: Gamification

- streak engine
- points engine
- badges
- rewards screen
- celebratory but restrained motion

Success criteria:

- daily use feels rewarding

## Phase 5: Intelligence

- anomaly detection
- opportunity recommendations
- provider optimization hints
- command optimization hints
- “easy day” recommendations

## Risks

- analytics depth may exceed current data collection
- provider/model pricing may need formal normalization
- layer-level metrics may require new instrumentation
- overdesign risk if we optimize for visuals before data correctness

## Non-Negotiables

- no placeholder metrics in shipped UI
- no giant empty cards
- no fake activity or decorative filler
- all savings and cost numbers must be explainable
- small terminal fallback must be explicit and clean

## Approval Gates

Before implementation starts, confirm:

1. Charm stack approved
2. dashboard-first product direction approved
3. V1 screens approved:
   - Home
   - Easy Day
   - Analytics
   - Providers
   - Models
   - Agents
   - Integrations
   - Logs
4. gamification in scope for Phase 3
5. cost model assumptions acceptable

## Recommendation

Approve a fresh rebuild with:

- Bubble Tea v2
- dashboard-first experience
- token + cost intelligence as the core product
- deep drilldowns after summary screens
- gamification after analytical foundation is correct
