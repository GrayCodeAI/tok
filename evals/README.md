# tok evals

Empirical measurement of compression quality and savings for the `tok` CLI and
the tok skill family.

## What's here

| File | Purpose |
|---|---|
| `bench.sh` | Shell benchmark: runs a fixture set through raw / tok / rtk and prints a tokens + ms table. No LLM. |
| `measure.py` | CI-side measurement reading snapshots produced by `llm_run.py`. |
| `llm_run.py` | Real-LLM snapshot generator. Requires `claude` CLI authenticated. |
| `semantic.py` | **Semantic drift harness** — does compression preserve downstream-answer correctness? Judge-model scored. |
| `semantic_fixtures.json` | Paired question+context fixtures for the semantic harness. |
| `plot.py` | Generates `snapshots/results.html` and `.png` boxplots from `results.json`. |
| `prompts/` | Prompt fixtures for the LLM eval matrix. |
| `snapshots/` | Output of `llm_run.py`. Checked in so CI is deterministic. |

## Running

```bash
# Quick shell benchmark (no LLM, no network)
./evals/bench.sh

# Full LLM eval (slow, costs tokens, requires authenticated claude CLI)
uv run python evals/llm_run.py

# Render plots from the latest snapshot
uv run --with tiktoken --with plotly --with kaleido python evals/plot.py
```

## CI hook

`measure.py` is the CI entry point. It reads `snapshots/results.json` and
fails the build if the median compression ratio regresses beyond a threshold.

A minimal GitHub Actions workflow:

```yaml
name: eval
on: [pull_request]
jobs:
  measure:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: "3.12" }
      - run: pip install tiktoken
      - run: python evals/measure.py
```

## Regenerating snapshots

Run `llm_run.py` locally (it calls the real Claude API). Commit the updated
`snapshots/results.json` in the same PR that changes skill content so reviewers
see the quality delta.

## Semantic drift harness

`semantic.py` is the separate harness that validates compression **preserves
downstream answer correctness**, not just token count. Each fixture pairs
a question with source context; the harness compresses the context via
`tok md --mode <mode>`, asks a model to answer using raw vs compressed
context, then uses a stricter judge model to decide if the answers are
materially equivalent.

```bash
export ANTHROPIC_API_KEY=sk-ant-...
python evals/semantic.py --mode ultra --threshold 0.9
python evals/semantic.py --mode wenyan-ultra --threshold 0.85
```

Without `ANTHROPIC_API_KEY`, the script prints what it would evaluate and
exits 0 — safe to run in CI jobs that don't have LLM credentials. Gate a
new compression mode on semantic drift staying above your threshold
before shipping it to users.
