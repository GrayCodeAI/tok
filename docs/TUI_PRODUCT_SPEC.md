# TokMan TUI Product Spec

## Purpose

This document is the implementation target for the new TokMan terminal UI.

It is stricter than the rebuild plan and should be treated as the product/source-of-truth document for the first TUI build.

It defines:

- screen set
- information architecture
- layout and interaction rules
- exact backend contracts
- V1 / V2 / V3 scope
- implementation order

The TUI should be built as a world-class terminal product for token reduction, token management, provider economics, and coding-agent efficiency.

## Product Definition

TokMan sits between:

- coding agent CLI/TUI
- IDE/editor/plugin integrations
- shell / hook / proxy / MCP execution
- downstream LLM provider/model usage

The TUI is the operator and executive view into that system.

It is not:

- a decorative dashboard
- a classic sysadmin console
- a generic logs pane
- a wrapper around individual CLI commands

It should answer:

- how many tokens TokMan reduced
- how much cost TokMan saved
- where savings came from
- where waste still exists
- which providers/models/agents/projects/sessions/commands matter most
- whether efficiency is improving or degrading
- what happened today

## Core Product Thesis

TokMan’s TUI should feel like a token intelligence cockpit.

Primary product pillars:

- token reduction analytics
- cost and provider/model intelligence
- coding-agent attribution
- session and command drilldown
- streaks, goals, and progress
- operational trust through diagnostics and data quality

Explicitly out of scope for the product surface:

- project dependency graph product screens
- cross-session memory product screens
- hosted/local LLM product screens

Those capabilities are not product pillars and should not shape the TUI architecture.

## Target Stack

- `Bubble Tea v2`
- `Bubbles v2`
- `Lip Gloss v2`
- `Glamour`
- `Huh` for setup/config flows when needed

Optional later:

- `Harmonica` for subtle, sparse motion only

## UX Principles

### 1. Dashboard First

The first screen must be useful in under 5 seconds.

### 2. Progressive Disclosure

Top-level screens summarize. Drilldowns explain.

### 3. Dense But Calm

High information density is required. Visual clutter is not.

### 4. Keyboard First

Everything important must be reachable quickly from the keyboard.

### 5. Stable Layout

The layout must not rely on fragile terminal width assumptions or oversized boxed regions.

### 6. Operational Trust

The product should always show data freshness, data quality, and integration health so users can trust what they are looking at.

## Primary User Jobs

### Executive Job

Answer:

- what did TokMan save today / this week / this month
- what did it cost before and after
- are we on pace or slipping
- what is the biggest win and the biggest problem

### Operator Job

Answer:

- which provider/model pairs are expensive
- which agents and commands are inefficient
- which sessions need inspection
- whether attribution/pricing/integration data is healthy

### Diagnostic Job

Answer:

- are hooks installed correctly
- is telemetry/store data healthy
- is pricing coverage complete enough
- are tracked commands missing critical attribution

## Information Architecture

Primary sections:

1. `Home`
2. `Easy Day`
3. `Analytics`
4. `Providers`
5. `Models`
6. `Agents`
7. `Sessions`
8. `Commands`
9. `Pipeline`
10. `Rewards`
11. `Logs`
12. `Config`

Recommended nav labels in the UI:

- `Home`
- `Today`
- `Trends`
- `Providers`
- `Models`
- `Agents`
- `Sessions`
- `Commands`
- `Pipeline`
- `Rewards`
- `Logs`
- `Config`

## Screen Specifications

### 1. Home

Purpose:

- overall token intelligence dashboard
- fastest path to “is TokMan helping?”

Must show:

- saved tokens in active window
- estimated cost saved
- reduction percent
- total commands
- active days in window
- streak summary
- points / level summary
- top provider
- top model
- top agent
- weakest command
- daily trend chart
- weekly trend chart
- daily budget summary
- data quality warning summary

Primary right-pane insights:

- top anomaly
- biggest gain
- most urgent warning
- best suggested next drilldown

Primary data:

- `tracking.GetWorkspaceDashboardSnapshot`

Success criteria:

- user can understand performance and health without leaving the screen

### 2. Easy Day

Purpose:

- answer “what happened today?” with minimal cognitive load

Must show:

- today saved tokens
- today estimated savings
- today command count
- today reduction percent
- today top provider/model
- today top project
- today top session
- best save today
- weakest pocket today
- today streak contribution
- one or two action-oriented insights

Format:

- fewer modules than Home
- larger, simpler KPI emphasis
- no secondary drilldown tables by default

Primary data:

- one-day `DashboardSnapshot`
- `DashboardDataQuality`

Success criteria:

- user can scan in 3-5 seconds and know whether today was good or bad

### 3. Analytics

Purpose:

- trend exploration beyond the executive summary

Must show:

- time-window switcher: `1d`, `7d`, `30d`, `90d`, `all`
- original vs filtered token trend
- saved token trend
- estimated savings trend
- reduction percent trend
- parse success trend
- budget utilization trend

UI model:

- chart-focused center pane
- metric toggles
- optional compact comparison table below charts

Primary data:

- `DashboardSnapshot`
- `DashboardDataQuality`

Success criteria:

- user can compare behavior across windows and spot trend changes

### 4. Providers

Purpose:

- provider-level cost and efficiency analysis

Must show:

- provider leaderboard by saved tokens
- provider leaderboard by cost saved
- provider leaderboard by original cost
- provider reduction percent
- provider-model pair leaderboard
- pricing coverage warning block
- fallback-priced model count

Drilldown detail:

- selected provider:
  - commands
  - sessions
  - projects
  - top models
  - efficiency notes

Primary data:

- `DashboardSnapshot.TopProviders`
- `DashboardSnapshot.TopProviderModels`
- `DashboardDataQuality.PricingCoverage`

Success criteria:

- user can see which providers cost the most and benefit the most

### 5. Models

Purpose:

- model-level economics and reduction analysis

Must show:

- model leaderboard
- saved tokens by model
- cost saved by model
- reduction percent by model
- fallback-priced / unknown models

Drilldown detail:

- selected model:
  - provider
  - command contribution
  - session contribution
  - estimated cost before/after

Primary data:

- `DashboardSnapshot.TopModels`
- `DashboardDataQuality.PricingCoverage`

Success criteria:

- user can identify costly or inefficient models immediately

### 6. Agents

Purpose:

- compare coding agents and integration channels

Must show:

- agent leaderboard by saved tokens
- agent leaderboard by cost saved
- agent command volume
- agent reduction percent
- integration health summary
- installed / partial / broken counts

Drilldown detail:

- selected agent:
  - providers/models used
  - top commands
  - projects/sessions
  - diagnostics summary

Primary data:

- `DashboardSnapshot.TopAgents`
- agent-aware `doctor`
- `tokman init --show`

Success criteria:

- user can compare agent efficiency and configuration quality

### 7. Sessions

Purpose:

- session-centric operational drilldown

Must show:

- recent sessions list
- total/active/snapshot counts
- top session agent
- active session context metrics
- snapshot history summary

Main list columns:

- session id
- agent
- project
- last activity
- turns
- tokens
- compression ratio
- snapshots
- active state

Detail pane:

- focus
- next action
- block type counts
- last snapshot time
- project path

Primary data:

- `session.GetAnalyticsSnapshot`
- `session.ListSessionOverviews`
- `session.ListSnapshotSummaries`
- `session.GetActiveContextMetrics`

Success criteria:

- user can inspect current and recent sessions without leaving the screen

### 8. Commands

Purpose:

- identify expensive and weak command patterns

Must show:

- top commands by saved tokens
- low-savings commands
- reduction percent by command
- cost impact by command
- parse failure hints

Main list columns:

- command
- commands count
- original tokens
- filtered tokens
- saved tokens
- reduction percent
- estimated savings

Detail pane:

- command quality summary
- likely waste pattern
- related provider/agent/project/session slices

Primary data:

- `DashboardSnapshot.TopCommands`
- `DashboardSnapshot.LowSavingsCommands`
- parse failure summary

Success criteria:

- user can quickly find which commands TokMan should optimize better

### 9. Pipeline

Purpose:

- explain where token reduction came from

Must show:

- layer leaderboard by total saved
- average saved per invocation
- contribution share
- low-impact layers

Main list columns:

- layer
- call count
- total saved
- average saved
- contribution %

Primary data:

- `DashboardSnapshot.TopLayers`
- `layer_stats`

Success criteria:

- user can understand which layers drive real value

### 10. Rewards

Purpose:

- progress, streaks, goals, motivation

Must show:

- savings streak
- goal streak
- points
- level
- next level progress
- badges
- best day
- best reduction day

Visual direction:

- slightly warmer accents than the rest of the product
- still restrained, not arcade-like

Primary data:

- `DashboardSnapshot.Streaks`
- `DashboardSnapshot.Gamification`

Success criteria:

- progress feels tangible without turning the app into a toy

### 11. Logs

Purpose:

- operational trace and trust

Must show:

- recent tracked operations
- warnings
- parse failures
- integration issues
- optional event severity filter

Primary sources:

- recent tracking history
- parse failure summary
- telemetry summary where useful
- doctor/config health summaries

Success criteria:

- user can troubleshoot without leaving the TUI

### 12. Config

Purpose:

- system state and configuration health

Must show:

- config path
- data path
- database path
- budget settings
- detected integrations
- doctor summary
- telemetry consent and store state
- integrity baseline state
- pricing coverage summary

Primary sources:

- config package
- `tokman doctor`
- `tokman init --show`
- `DashboardDataQuality`

Success criteria:

- user can explain product state and setup health from one screen

## Layout Model

### Standard Layout

Use for terminals at or above a comfortable width/height threshold.

Structure:

- top header bar
- left navigation rail
- center content pane
- right context/insight pane
- bottom shortcut/status bar

The right pane should collapse automatically when it does not add value.

### Compact Layout

Use on smaller terminals.

Structure:

- header
- horizontal section tabs
- single main content pane
- bottom shortcut/status bar

Behavior:

- insight pane becomes overlay-only
- dense tables show fewer columns
- some charts fall back to summary bands

### Overlay Policy

Fullscreen overlays allowed for:

- help
- search
- filter menu
- command detail
- session detail
- long logs

Do not use overlays for routine navigation.

## Interaction Model

### Global Keys

- `1-9` jump to primary sections
- `0` or mnemonic letter for later sections if needed
- `tab` / `shift+tab` move focus zones
- `j` / `k` or arrows move selection
- `enter` open detail / apply
- `esc` go back / close overlay
- `/` search
- `f` filter
- `s` sort
- `w` cycle time window
- `r` refresh
- `g` go to section picker
- `?` help
- `q` quit

### Focus Zones

There should be only a few major focus zones:

- navigation
- primary content
- context pane
- overlay

Users must always know which zone is active.

### Search Model

Search should support:

- section-local search
- fuzzy match against current list/table
- immediate narrowing, not a separate modal workflow by default

### Filter Model

Global filters should be consistent:

- time window
- project
- agent
- provider
- model
- session

Not every screen needs every filter visible all the time, but the state model should support all of them.

## Data Contracts

### Root Aggregate

Preferred root payload:

- [internal/tracking/workspace_snapshot.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/tracking/workspace_snapshot.go:1)

This combines:

- `Dashboard`
- `DataQuality`
- `Sessions`

### Dashboard Aggregate

Use:

- [internal/tracking/dashboard.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/tracking/dashboard.go:1)

Important fields:

- `Overview`
- `DailyTrends`
- `WeeklyTrends`
- `TopAgents`
- `TopProviders`
- `TopModels`
- `TopProviderModels`
- `TopProjects`
- `TopCommands`
- `TopSessions`
- `ContextKinds`
- `TopLayers`
- `LowSavingsCommands`
- `Budgets`
- `Streaks`
- `Lifecycle`
- `Gamification`

### Data Quality Aggregate

Use:

- [internal/tracking/data_quality.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/tracking/data_quality.go:1)

Important fields:

- total commands
- missing agent/provider/model/session attribution
- parse failures
- pricing coverage
- fallback-priced models
- unknown model names

### Session Aggregate

Use:

- [internal/session/types.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/session/types.go:77)
- [internal/session/manager.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/session/manager.go:515)

Important fields:

- store summary
- recent sessions
- snapshot history
- active session context

## State Model

The TUI should have explicit state for:

- current section
- terminal size
- compact vs standard layout
- selected row/item per section
- search query
- active filters
- sort field + direction
- refresh state
- loading state
- last successful snapshot time
- last error state
- overlay state

Suggested root model slices:

- `appState`
- `themeState`
- `dataState`
- `filterState`
- `navState`
- `overlayState`

## Refresh Model

Default:

- manual refresh available always
- passive auto-refresh on a moderate interval

Recommended default intervals:

- `Home` / `Easy Day`: 15-30s
- `Analytics` / `Providers` / `Models` / `Agents`: 30-60s
- `Sessions` / `Logs`: 10-15s when active
- `Config`: manual or slow refresh

Rules:

- no aggressive redraw churn
- refresh indicator must be visible
- stale data state must be obvious

## Chart Model

Use simple terminal-native visualizations:

- sparkline or bandline for trends
- compact bars for leaderboards
- delta arrows for trend direction
- progress bars for budgets and level progress

Do not build:

- pseudo-3D charts
- heavy ASCII art dashboards
- over-animated graphs

Charts should support:

- current value
- previous value
- direction
- min/max where useful

## Visual Direction

### Tone

- premium
- technical
- modern
- restrained

### Structural Rules

- prefer spacing and alignment over full borders
- use dividers sparingly
- use color semantically, not decoratively
- tables should be readable first, stylish second
- warnings must stand out immediately

### Semantic Colors

- savings / healthy efficiency: green
- cost pressure / warnings: amber
- failures / broken integrations: red
- neutral structure: slate / dim gray
- active selection / focus: cyan or blue
- rewards / streaks: gold accents

### Typography Rules

- one strong header style
- one section title style
- one muted metadata style
- one danger/warning style

Do not create too many competing text treatments.

## Small Terminal Behavior

At small sizes:

- collapse right pane
- hide low-value metadata
- reduce table columns
- switch charts to compact summaries
- keep shortcuts visible
- preserve navigation clarity over completeness

If a screen becomes too dense:

- show the important subset
- offer detail via overlay

Never render giant empty framed areas.

## Empty / Error / Loading States

### Empty

Must explain:

- whether there is no data
- whether filters caused the empty result
- what action to take next

### Error

Must show:

- what failed
- whether retry is possible
- whether cached/stale data is being shown

### Loading

Must be restrained:

- single status line or compact placeholder rows
- avoid shimmer-heavy behavior

## V1 Scope

Build first:

- app shell
- nav model
- filter model
- refresh model
- Home
- Easy Day
- Analytics
- Providers
- Agents
- Sessions
- Commands
- Rewards
- Config
- help / search / filter overlay shell

V1 should not depend on heavy motion or exotic charting.

## V2 Scope

Add next:

- fuller Models screen
- stronger Pipeline screen
- richer Logs screen
- export/report overlays
- comparative analytics
- better budget coaching
- more anomaly hints

## V3 Scope

Later only:

- multi-project switching at scale
- team/shared leaderboards
- predictive pacing
- anomaly detection beyond simple thresholds
- background workers/materialized dashboard caches

## Implementation Phases

### Phase 1: Shell

- app skeleton
- theme tokens
- nav system
- responsive layout engine
- screen routing
- overlay system

### Phase 2: Data Layer Integration

- workspace snapshot loader
- refresh loop
- loading/error/empty states
- filter state + serialization

### Phase 3: Core Screens

- Home
- Easy Day
- Providers
- Sessions
- Commands
- Rewards
- Config

### Phase 4: Secondary Screens

- Analytics
- Agents
- Models
- Pipeline
- Logs

### Phase 5: Drilldowns and Ergonomics

- detail overlays
- search/filter/sort polish
- keyboard ergonomics
- compact terminal tuning

### Phase 6: Product Polish

- spacing/contrast pass
- chart legibility pass
- consistency pass
- performance pass

## Build Gates

Before implementation starts:

- do not draw decorative containers first
- do not invent new backend fields without checking existing contracts
- do not let the layout depend on wide terminals
- do not build from ad hoc queries

The TUI should be built on:

- [internal/tracking/dashboard.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/tracking/dashboard.go:1)
- [internal/tracking/data_quality.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/tracking/data_quality.go:1)
- [internal/tracking/workspace_snapshot.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/tracking/workspace_snapshot.go:1)
- [internal/session/manager.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/session/manager.go:515)
- agent-aware diagnostics from `tokman doctor`
- real integration state from `tokman init --show`

## Approval Target

Once this spec is accepted, the next implementation step should be:

1. create the Bubble Tea shell
2. implement the shared layout system
3. wire one real data loader
4. build `Home` and `Easy Day` first
