# Tok Test Runner

The `tok test-runner` command provides automatic test runner detection and execution with intelligent output filtering.

## Overview

Test-runner automatically detects the appropriate test framework for your project and runs tests with optimized, token-efficient output. It supports 10+ test frameworks out of the box.

## Features

- **Auto-detection**: Automatically identifies test frameworks based on project files
- **Multi-framework support**: Works with Cargo, Go, npm, pnpm, Pytest, Vitest, Jest, RSpec, Rake, and Playwright
- **Smart filtering**: Reduces test output by 60-90% while preserving failure details
- **Priority-based selection**: More specific configurations are preferred (e.g., Vitest over npm)
- **Integration**: Works with the auto-rewrite system for transparent usage

## Supported Test Frameworks

| Framework | Detection File(s) | Priority |
|-----------|-------------------|----------|
| **Vitest** | `vitest.config.ts/js` | 110 (highest) |
| **Playwright** | `playwright.config.ts/js` | 105 |
| **Cargo** | `Cargo.toml` | 100 |
| **Go** | `go.mod` | 100 |
| **RSpec** | `.rspec`, `spec/` | 100 |
| **Pytest** | `pytest.ini`, `setup.py` | 100 |
| **pnpm** | `pnpm-lock.yaml` | 75 |
| **npm** | `package.json` | 70 |
| **Jest** | `jest.config.js/ts` | 80 |
| **Rake** | `Rakefile` | 80 |

## Usage

### Auto-detect and run tests

```bash
# Automatically detect and run tests
tok test-runner

# In a Rust project
cd my-rust-project
tok test-runner
# Output: Running Cargo tests...
#         ✓ 15 tests passed

# In a Node.js project
cd my-node-project
tok test-runner
# Output: Running npm tests...
#         ✓ 42 tests passed
```

### Explicit test runner

```bash
# Specify the test command explicitly
tok test-runner cargo test
tok test-runner npm test
tok test-runner pytest -v
tok test-runner go test ./...
```

### With auto-rewrite

When using Tok's auto-rewrite hook, test commands are automatically converted:

```bash
# These are automatically rewritten:
cargo test           → tok test-runner cargo test
npm test             → tok test-runner npm test
pytest               → tok test-runner pytest
go test ./...        → tok test-runner go test ./...
```

## Output Filtering

Test-runner intelligently filters test output to reduce token usage:

### What's Included

- Test failure details
- Stack traces (first 5 lines)
- Summary statistics (pass/fail counts)
- Error messages

### What's Filtered

- Passing test details (unless verbose mode)
- Progress bars
- Unnecessary whitespace
- Duplicate information

### Example Output

**Before (500+ lines):**
```
running 100 tests
test test_001 ... ok
test test_002 ... ok
... (96 more passing tests)
test test_099 ... FAILED
thread 'test_099' panicked at 'assertion failed', src/lib.rs:42:5
note: run with RUST_BACKTRACE=1 environment variable
failures:
    test_099
test result: FAILED. 98 passed; 1 failed; 0 ignored
```

**After (10 lines):**
```
Running Cargo tests...
FAILED: 1/100 tests

Test Failures:
────────────────────────────────────────
test test_099: assertion failed
  at src/lib.rs:42:5

✗ 98 passed, 1 failed
```

## Configuration

### Disable auto-rewrite for specific commands

Add to your `~/.config/tok/config.toml`:

```toml
[hooks]
excluded_commands = ["cargo test", "npm test"]
```

### Prefer explicit tok commands

```toml
[rewrite]
prefer_explicit = true  # Use "tok cargo test" instead of "tok test-runner cargo test"
```

## Integration with CI/CD

Test-runner works great in CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Run tests
  run: tok test-runner

- name: Run tests with coverage
  run: tok test-runner cargo test --coverage
```

## Performance

- **Detection speed**: < 1ms for project scanning
- **Filtering overhead**: < 5ms for typical test output
- **Token reduction**: 60-90% on average

## Troubleshooting

### No test runner detected

If no test runner is detected:

1. Check that the project has the appropriate configuration file
2. Run with verbose mode: `tok -v test-runner`
3. Explicitly specify the test command: `tok test-runner cargo test`

### Wrong test runner selected

If the wrong test runner is auto-detected:

1. Check the priority table above
2. Use explicit mode: `tok test-runner <correct-command>`
3. Adjust priority by removing conflicting configuration files

### Missing test failures

If failures aren't showing in output:

1. Run with verbose mode: `tok -vv test-runner`
2. Check the raw output: `tok -vvv test-runner`
3. Report an issue with the test framework and output format

## Examples

### Rust project with Cargo

```bash
$ cd my-rust-project
$ tok test-runner
Running Cargo tests...
✓ 47 tests passed (12ms)
```

### Node.js project with Vitest

```bash
$ cd my-vite-project
$ tok test-runner
Running Vitest...
✓ 156 tests passed
```

### Python project with Pytest

```bash
$ cd my-python-project
$ tok test-runner
Running pytest...
FAILED: 2/50 tests

Test Failures:
────────────────────────────────────────
test_api.py::test_endpoint: Connection refused
  at tests/test_api.py:23:4

test_models.py::test_validation: AssertionError
  at tests/test_models.py:45:8

✗ 48 passed, 2 failed
```

## See Also

- [Auto-Rewrite System](./AGENT_INTEGRATION.md)
- [Configuration Guide](./TUNING.md)
- [Benchmarks](./BENCHMARKS.md)
