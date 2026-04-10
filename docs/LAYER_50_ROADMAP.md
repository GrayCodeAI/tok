# TokMan 50-Layer Core Roadmap

Date: April 10, 2026

Goal: build a world-class token optimization engine with 50+ research-backed layers while preserving quality and predictable latency.

## Core Principles

1. Never run all 50 layers on every request.
2. Route dynamically by command family, intent, size, and budget.
3. Keep quality guardrails mandatory for aggressive paths.
4. Promote layers using benchmark + ablation evidence only.

## 50-Layer Architecture

1. `pre` stages (5): coarse reduction and signal retention.
2. `core` stages (15): stable high-value layers for general workloads.
3. `adaptive` stages (15): conditional layers for task-specific optimization.
4. `recovery` stages (10): safety and recall repair layers.
5. `post` stages (5): budget polishing, readability, and output normalization.

Current implementation status:
- Implemented layer metadata + tiering in [layer_registry.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/filter/layer_registry.go)
- Gate framework in [layer_gate.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/filter/layer_gate.go)
- Pipeline execution hook in [pipeline_gates.go](/Users/lakshmanpatel/Desktop/ProjectAlpha/tokman/internal/filter/pipeline_gates.go)

## Layer Gate Modes

1. `all` (default): preserve current behavior and run all enabled layers.
2. `stable-only`: run stable/recovery tiers and skip experimental layers unless allow-listed.

Config fields:
- `LayerGateMode`
- `LayerGateAllowExperimental`

## Promotion Policy (Experimental -> Stable)

A layer is promoted only if:

1. It improves average token savings on benchmark suite.
2. It does not regress quality guardrail retention metrics.
3. p95 latency remains within target envelope.
4. Ablation shows positive marginal value.

## Benchmark and Ablation Plan

1. Run baseline vs adaptive comparisons.
2. Run scenario suite (build/test/diff/ops/multi-turn).
3. Run per-layer ablation to estimate marginal gain.
4. Store artifacts in CI and compare over time.

## Implementation Milestones

1. M1: Registry + gate framework (completed).
2. M2: Expand planned layers 30-49 with experimental implementations.
3. M3: Add per-command-family calibrated routing.
4. M4: Add automatic retirement for low-value layers.

