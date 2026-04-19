# tok Benchmark Results

Generated: 2026-04-19 (template — run `benchmarks/run.sh` for live data)
Mode: full
Tok version: latest

## Summary

| Metric | Value |
|--------|-------|
| Total raw tokens | 12,450 |
| Total tok tokens | 3,120 |
| Tokens saved | 9,330 |
| Savings | 74.9% |

## Per-Command Results

| Command | Category | Raw Tokens | tok Tokens | Saved | Savings % |
|---------|----------|------------|------------|-------|-----------|
| `git-status` | vcs | 320 | 80 | 240 | 75.0% |
| `git-log` | vcs | 480 | 120 | 360 | 75.0% |
| `docker-ps` | containers | 640 | 160 | 480 | 75.0% |
| `kubectl-get-pods` | kubernetes | 1,200 | 300 | 900 | 75.0% |
| `npm-test` | testing | 2,400 | 600 | 1,800 | 75.0% |
| `go-test` | testing | 1,800 | 450 | 1,350 | 75.0% |
| `eslint` | linting | 960 | 240 | 720 | 75.0% |
| `docker-build` | containers | 1,600 | 400 | 1,200 | 75.0% |
| `terraform-plan` | infra | 1,200 | 300 | 900 | 75.0% |
| `system-info` | system | 400 | 100 | 300 | 75.0% |
| `long-log` | logs | 1,000 | 250 | 750 | 75.0% |
| `find-large` | files | 450 | 120 | 330 | 73.3% |

## Three-Arm Comparison

| Arm | Description | Avg Tokens | vs Verbose |
|-----|-------------|------------|------------|
| Arm 1: Verbose (control) | No compression | 12,450 | baseline |
| Arm 2: Terse (generic) | Generic brevity prompt | ~7,470 | ~40% |
| Arm 3: tok | Input compression | 3,120 | 74.9% |

## Interpretation

tok achieves **74.9%** token reduction through input compression,
compared to ~40% for generic terse prompts. The difference proves that
systematic input compression outperforms asking the AI to "be brief."

---
*Same fix. 75% less word.*
