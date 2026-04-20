# End-to-end tests (Docker-based)

Runs `tok` against real CLIs (`git`, `kubectl`, `docker`, `aws` stubs) inside
disposable containers so we verify the happy-path of every wrapper without
touching the host. Docker stand-in for the Multipass VM suite that rtk uses.

## Structure

```
tests/e2e/
├── README.md        (this file)
├── run.sh           driver: builds images, runs every case, tallies pass/fail
├── fixtures/        per-scenario Dockerfiles + input scripts
│   ├── git/
│   │   ├── Dockerfile
│   │   └── case.sh
│   ├── kubectl/
│   ├── aws/
│   └── ...
└── golden/          expected-output snapshots (for byte-identical checks)
```

## Conventions

- Each scenario is a self-contained directory under `fixtures/`.
- Each scenario has a `Dockerfile` installing the target CLI + a sibling
  `case.sh` that invokes it and prints the output.
- `run.sh` builds the image, runs the case, captures stdout, pipes it through
  `tok` (host binary mounted at `/tok`), and diffs against `golden/<name>.out`.

## Running

```bash
# Build tok, then run the suite:
go build -o tok ./cmd/tok
tests/e2e/run.sh                # all scenarios
tests/e2e/run.sh git kubectl    # subset
TOK_E2E_UPDATE=1 tests/e2e/run.sh   # rewrite golden files in place
```

## Why not Multipass?

Multipass (rtk's choice) is macOS-friendly and boots in seconds on ARM, but
assumes a hypervisor. Docker is more portable across CI runners and avoids
nested virtualization. Scenarios needing a full kernel (`mount`, `iptables`)
can be marked `requires: privileged` and skipped on restricted CI.

## CI wiring (future)

Planned `.github/workflows/e2e.yml`:

```yaml
name: e2e
on: [pull_request]
jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "stable" }
      - run: go build -o tok ./cmd/tok
      - run: tests/e2e/run.sh
```
