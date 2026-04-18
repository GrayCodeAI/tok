# TokMan Dashboard Data Layer

## Purpose

This document defines the canonical backend aggregate surface for the future TUI.

The goal is to stop building dashboard views from ad hoc SQL scattered across commands and instead expose one stable tracking API that the UI can trust.

## Canonical API

Implemented in:

- `internal/tracking/dashboard.go`

Primary entry point:

- `(*Tracker).GetDashboardSnapshot(opts DashboardQueryOptions)`

Supporting aggregate methods:

- `GetDashboardOverview`
- `GetDashboardTrends`
- `GetDashboardBreakdown`
- `GetDashboardTopLayers`
- `GetDashboardLowSavingsCommands`
- `GetDashboardBudgets`
- `GetDashboardStreaks`
- `GetDashboardLifecycle`

## Supported Filters

`DashboardQueryOptions` currently supports:

- `Days`
- `ProjectPath`
- `AgentName`
- `Provider`
- `ModelName`
- `SessionID`
- `Limit`
- `ReductionGoalPct`
- `DailyTokenBudget`
- `WeeklyTokenBudget`
- `MonthlyTokenBudget`
- `DailyCostBudgetUSD`
- `WeeklyCostBudgetUSD`
- `MonthlyCostBudgetUSD`

Default behaviors:

- `Days <= 0` becomes `30`
- `Limit <= 0` becomes `10`
- `Limit > 100` is capped to `100`

## Snapshot Shape

`DashboardSnapshot` currently exposes:

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

## Breakdown Dimensions

The canonical breakdown API supports these dimensions:

- `agent`
- `provider`
- `model`
- `provider_model`
- `project`
- `command`
- `session`
- `context_kind`

Unknown or empty values are normalized to:

- `"(unknown)"`

## Cost Semantics

Cost fields are estimated from tracked token counts plus TokMan's current model pricing table in `internal/tracking/cost.go`.

Current cost fields:

- `EstimatedOriginalCostUSD`
- `EstimatedFilteredCostUSD`
- `EstimatedSavingsUSD`

Important limitation:

- pricing coverage is still only as good as the current model pricing table
- unsupported or unknown models fall back to TokMan's default estimator

## Current Scope

This data layer is based on the `commands` tracking database plus `layer_stats`.

That means it already supports:

- provider analytics
- model analytics
- agent analytics
- command analytics
- project analytics
- session analytics from tracked `session_id`
- context-kind analytics
- layer effectiveness
- low-savings command detection
- budget windows for token management
- savings and efficiency streaks
- lifecycle metrics such as first seen date, active days, and tracked project count
- gamification points and badges

## Not Yet Included

These areas still need expansion before the TUI is complete:

- composition with `sessions.db` for richer session metadata
- IDE/editor attribution as a first-class dimension
- provider/model pricing normalization beyond the current pricing table
- pre-aggregated dashboard cache/materialized views for very large histories

## Usage Direction

The TUI should consume `DashboardSnapshot` first for the main dashboard and only fall back to narrower aggregate methods for drilldowns, filters, or detail views.
