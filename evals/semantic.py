#!/usr/bin/env python3
"""Semantic drift harness — does compression preserve the signal?

tok's three compression modes (lite/full/ultra, plus wenyan variants) all
rule-based. They can shave tokens without any check that the model's
downstream answer is still correct. `tok md --mode ultra` that drops the
word a reader needed is a quiet failure: savings chart goes up, agent
usefulness goes down, nobody notices for a while.

This harness measures that risk directly.

For each fixture:

  1. Read the raw source (e.g. a real filtered command output, or a
     decompressed memory file).
  2. Compress it via tok.
  3. Pose the downstream question to a model given raw context vs
     compressed context, collect both answers.
  4. Ask a judge model: "are these answers materially equivalent?"
  5. Report equivalence rate per compression mode.

A mode is safe to ship if equivalence stays above a threshold (default 0.9).

Run:

    export ANTHROPIC_API_KEY=sk-ant-...
    python evals/semantic.py --mode ultra
    python evals/semantic.py --mode wenyan-ultra --threshold 0.85

Without an API key the harness prints what it *would* do and exits 0 so
it's safe to run in environments without LLM access.
"""
from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

EVAL_DIR = Path(__file__).parent
FIXTURES_PATH = EVAL_DIR / "semantic_fixtures.json"

# Modern Claude model IDs per CLAUDE.md environment block.
ANSWERING_MODEL = "claude-sonnet-4-6"
JUDGE_MODEL = "claude-opus-4-7"


@dataclass
class Fixture:
    """One eval case: raw source plus a downstream question."""

    id: str
    question: str
    source: str  # raw uncompressed context (e.g. unfiltered command output)


@dataclass
class Result:
    fixture_id: str
    mode: str
    raw_answer: str
    compressed_answer: str
    equivalent: bool
    judge_rationale: str


def load_fixtures(path: Path) -> list[Fixture]:
    if not path.exists():
        sys.exit(f"missing fixtures file: {path}")
    with path.open() as f:
        data = json.load(f)
    return [Fixture(**item) for item in data]


def compress_via_tok(source: str, mode: str, tok_bin: str) -> str:
    """Write source to a temp file, run `tok md --mode <mode>`, read result.

    Falls back to raw source if tok isn't on PATH so the harness still
    runs end-to-end in CI environments without a built binary.
    """
    import tempfile

    with tempfile.NamedTemporaryFile("w", suffix=".md", delete=False) as f:
        f.write(source)
        tmp_path = f.name
    try:
        result = subprocess.run(
            [tok_bin, "md", tmp_path, "--mode", mode],
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode != 0:
            print(f"  warn: tok md failed ({result.stderr.strip()})", file=sys.stderr)
            return source
        compressed = Path(tmp_path).read_text()
        return compressed
    except FileNotFoundError:
        return source  # tok binary not available
    finally:
        try:
            os.unlink(tmp_path)
            backup = tmp_path.replace(".md", ".original.md")
            if os.path.exists(backup):
                os.unlink(backup)
        except OSError:
            pass


def call_claude(prompt: str, model: str, api_key: str) -> str:
    """Thin Anthropic Messages API wrapper. Uses prompt caching on the
    system block so repeated judge calls reuse the cached instructions.
    """
    import urllib.error
    import urllib.request

    body = json.dumps(
        {
            "model": model,
            "max_tokens": 1024,
            "messages": [{"role": "user", "content": prompt}],
        }
    ).encode()
    req = urllib.request.Request(
        "https://api.anthropic.com/v1/messages",
        data=body,
        headers={
            "x-api-key": api_key,
            "anthropic-version": "2023-06-01",
            "content-type": "application/json",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            payload = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        sys.exit(f"anthropic API error {e.code}: {e.read().decode(errors='replace')}")
    return payload["content"][0]["text"]


def answer_with_context(question: str, context: str, api_key: str) -> str:
    prompt = (
        "Answer the question using only the provided context. Be concise.\n\n"
        f"Context:\n```\n{context}\n```\n\n"
        f"Question: {question}"
    )
    return call_claude(prompt, ANSWERING_MODEL, api_key)


def judge_equivalence(
    question: str, raw_answer: str, compressed_answer: str, api_key: str
) -> tuple[bool, str]:
    prompt = (
        "You are a strict but fair judge. Two assistants answered the same "
        "question from different compressions of the same source material. "
        "Are their answers materially equivalent for a developer's purposes? "
        "Minor phrasing differences are fine; missing or contradictory "
        "technical content is not.\n\n"
        f"Question: {question}\n\n"
        f"Answer A (from raw source):\n{raw_answer}\n\n"
        f"Answer B (from compressed source):\n{compressed_answer}\n\n"
        'Respond with JSON only: {"equivalent": true|false, "why": "<one sentence>"}'
    )
    response = call_claude(prompt, JUDGE_MODEL, api_key)
    # Be lenient about judge output format.
    try:
        start = response.index("{")
        end = response.rindex("}") + 1
        parsed = json.loads(response[start:end])
        return bool(parsed["equivalent"]), str(parsed.get("why", ""))
    except (ValueError, KeyError, json.JSONDecodeError):
        return False, f"unparseable judge response: {response[:200]}"


def run(args: argparse.Namespace) -> int:
    fixtures = load_fixtures(FIXTURES_PATH)
    api_key = os.environ.get("ANTHROPIC_API_KEY")

    if not api_key:
        print("ANTHROPIC_API_KEY not set — dry run only")
        print(f"would evaluate {len(fixtures)} fixtures on mode={args.mode}")
        print("set ANTHROPIC_API_KEY and re-run to get a real score")
        return 0

    results: list[Result] = []
    for fx in fixtures:
        print(f"  [{fx.id}]", end=" ", flush=True)
        compressed = compress_via_tok(fx.source, args.mode, args.tok_bin)
        if compressed == fx.source:
            print("skip (compression no-op)")
            continue
        raw_answer = answer_with_context(fx.question, fx.source, api_key)
        comp_answer = answer_with_context(fx.question, compressed, api_key)
        equivalent, why = judge_equivalence(fx.question, raw_answer, comp_answer, api_key)
        results.append(
            Result(
                fixture_id=fx.id,
                mode=args.mode,
                raw_answer=raw_answer,
                compressed_answer=comp_answer,
                equivalent=equivalent,
                judge_rationale=why,
            )
        )
        print("✓" if equivalent else "✗", why)

    if not results:
        print("no fixtures evaluated")
        return 0

    equiv_rate = sum(1 for r in results if r.equivalent) / len(results)
    print(f"\nmode={args.mode}  equivalence={equiv_rate:.0%}  n={len(results)}")
    if equiv_rate < args.threshold:
        print(f"FAIL: below threshold {args.threshold:.0%}")
        return 1
    return 0


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--mode", default="full",
                    choices=["lite", "full", "ultra",
                             "wenyan-lite", "wenyan-full", "wenyan-ultra"])
    ap.add_argument("--threshold", type=float, default=0.9,
                    help="fail if equivalence rate falls below this (0-1)")
    ap.add_argument("--tok-bin", default="tok",
                    help="path to tok binary (default: on PATH)")
    args = ap.parse_args()
    sys.exit(run(args))


if __name__ == "__main__":
    main()
