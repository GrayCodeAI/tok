# TokMan Research Roadmap (2026)

This roadmap translates recent long-context and prompt-compression research into concrete TokMan engineering milestones.

Date: April 10, 2026

## Research Signals

1. Extractive compression is consistently strong for quality/cost tradeoffs.
   - LLMLingua (EMNLP 2023): https://aclanthology.org/2023.emnlp-main.391/
   - Selective Context (2023): https://arxiv.org/abs/2310.06201
2. Query/task-conditioned compression improves relevance for coding tasks.
   - LongLLMLingua (2024): https://arxiv.org/abs/2310.06839
   - SWE-Pruner (2026): https://arxiv.org/abs/2601.16746
3. New token-classification compressors reduce latency/overhead.
   - LLMLingua-2 (2024): https://arxiv.org/abs/2403.12968
4. Long-session quality depends on cache lifecycle, not one-pass metrics.
   - SCBench (2024/2025): https://arxiv.org/abs/2412.10319
5. Streaming/KV strategies remain high-impact for long contexts.
   - H2O (2023): https://arxiv.org/abs/2306.14048
   - StreamingLLM (2023): https://arxiv.org/abs/2309.17453
   - SCA (2024): https://aclanthology.org/2024.findings-emnlp.358/

## Product Goals

1. Improve task-success retention at fixed token budget.
2. Reduce unnecessary layer execution for lower latency.
3. Add benchmark coverage for multi-turn and long-session behavior.
4. Keep behavior deterministic and explainable in production.

## Milestone Plan

## M1: Extractive-First Path (P0)

Status: In progress (scaffold merged)

Deliverables:
- Optional extractive prefilter before deep pipeline layers.
- Preserve head/tail + high-signal error/failure lines.
- Layer stats key: `pre_extractive`.

Acceptance:
- >= 20% latency reduction on very large logs (> 400 lines).
- No task success drop > 2% on CLI regression set.

## M2: Policy Router (P0)

Status: In progress (scaffold merged)

Deliverables:
- Optional routing stage that infers query intent from raw output.
- Auto-enable query-dependent layers when safe.

Acceptance:
- Better retention for debugging/test workflows in eval set.
- No degradation for explicit user intent inputs.

## M3: Benchmark Suite 2.0 (P0)

Status: Planned

Deliverables:
- Scenario packs:
  - single-shot command output compression
  - long test logs with repeated failures
  - multi-turn shared context replay
  - diff-heavy code review context
- Metrics:
  - token reduction
  - wall-clock latency
  - task-success proxy (contains critical lines)
  - omission/regression rate

Acceptance:
- CI benchmark report on every PR touching `internal/filter/*`.

## M4: Adaptive Layer Router (P1)

Status: Planned

Deliverables:
- Route by content type, estimated entropy, length, and intent.
- Skip low-value layers for short/sparse outputs.

Acceptance:
- p95 latency improvement with equal or better quality score.

## M5: Quality Guardrails + Auto-Fallback (P1)

Status: Planned

Deliverables:
- Post-compression checks for critical markers:
  - error/stack traces
  - changed-file hunks
  - test failures and assertions
- Automatic fallback to safer profile when checks fail.

Acceptance:
- Significant drop in harmful omissions vs baseline.

## M6: Layer Ablation and Retirement (P2)

Status: Planned

Deliverables:
- Measure marginal contribution of each layer by command family.
- Remove or demote low-yield layers.

Acceptance:
- Simpler pipeline with lower complexity and equal quality.

## Implementation Notes

1. Keep all new features behind config flags first.
2. Add tests for both pass-through and enabled modes.
3. Track each milestone with benchmark artifacts in CI.
4. Document every layer decision to prevent drift between code and docs.
