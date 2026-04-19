#!/usr/bin/env python3
import subprocess

SAMPLES = [
    "Please review this authentication middleware and suggest a fix for token expiry handling.",
    "I would be happy to help you set up a PostgreSQL connection pool for the service.",
    "Could you explain why this React component keeps re-rendering when props look unchanged?",
]

MODES = ["lite", "full", "ultra"]


def est_tokens(text: str) -> int:
    return max(1, len(text) // 4)


def run_mode(mode: str, text: str) -> str:
    result = subprocess.run(
        ["go", "run", "./cmd/tok", "input", "compress", "-mode", mode, "-input", text],
        capture_output=True,
        text=True,
        check=True,
    )
    return result.stdout.strip()


def main() -> None:
    print("| mode | original | compressed | saved |")
    print("|------|----------|------------|-------|")
    for mode in MODES:
        orig_total = 0
        comp_total = 0
        for sample in SAMPLES:
            orig = est_tokens(sample)
            comp = est_tokens(run_mode(mode, sample))
            orig_total += orig
            comp_total += comp
        saved = 100.0 * (orig_total - comp_total) / float(orig_total)
        print(f"| {mode} | {orig_total} | {comp_total} | {saved:.1f}% |")


if __name__ == "__main__":
    main()
