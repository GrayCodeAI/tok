# Custom TOML Filters

## Overview

TokMan uses TOML filter files to define how specific commands should be compressed. You can create custom filters for your own tools and workflows.

## Filter Location

```
~/.config/tokman/filters/
├── git.toml          # Built-in git filters
├── go.toml           # Built-in go filters
├── npm.toml          # Built-in npm filters
├── my-tool.toml      # Your custom filter
└── deploy.toml       # Your deploy tool filter
```

## Basic Filter

```toml
# ~/.config/tokman/filters/my-tool.toml

[my_tool_build]
match = "^my-tool build"
description = "Compress my-tool build output"
strip_lines_matching = [
    "^\\[info\\]",           # Remove info lines
    "^Downloading",          # Remove download progress
    "^\\s*$",                # Remove empty lines
]
max_lines = 30
```

## Advanced Filter

```toml
# ~/.config/tokman/filters/deploy.toml

[deploy_production]
match = "^deploy (prod|production)"
description = "Compress production deployment output"
strip_lines_matching = [
    "^\\[DEBUG\\]",
    "^Connecting to",
    "^Uploading artifact",
    "^Waiting for",
]
output_patterns = [
    "Deployment successful",
    "Deployment failed",
    "ERROR:",
    "WARNING:",
]
max_lines = 20

[deploy_staging]
match = "^deploy (staging|stage)"
description = "Compress staging deployment output"
strip_lines_matching = [
    "^\\[DEBUG\\]",
    "^\\[TRACE\\]",
]
max_lines = 50
```

## Filter with Grouping

```toml
# ~/.config/tokman/filters/webpack.toml

[webpack_build]
match = "^(npx |)webpack"
description = "Compress webpack build output"
strip_lines_matching = [
    "^\\s*asset ",           # Individual asset lines
    "^\\s*modules ",         # Module count lines
    "^\\s*\\+ \\d+ modules", # Hidden modules
]
output_patterns = [
    "ERROR in",
    "WARNING in",
    "compiled successfully",
    "compiled with \\d+ error",
    "compiled with \\d+ warning",
]
max_lines = 25
```

## Filter for Test Runners

```toml
# ~/.config/tokman/filters/custom-test.toml

[my_test_runner]
match = "^my-test-runner"
description = "Custom test runner output compression"
strip_lines_matching = [
    "^\\s*✓",               # Passing tests
    "^\\s*PASS",            # Pass markers
    "^\\s*Running",         # Progress indicators
]
# Keep only failures and summary
output_patterns = [
    "FAIL",
    "ERROR",
    "\\d+ passed",
    "\\d+ failed",
    "Total:",
]
max_lines = 40
```

## Filter for Log Parsers

```toml
# ~/.config/tokman/filters/logs.toml

[tail_app_log]
match = "^tail.*app\\.log"
description = "Compress application log output"
strip_lines_matching = [
    "^\\[INFO\\]",
    "^\\[DEBUG\\]",
    "^\\[TRACE\\]",
    "^\\s*at ",             # Stack trace internals
]
# Keep errors and warnings
output_patterns = [
    "\\[ERROR\\]",
    "\\[WARN\\]",
    "\\[FATAL\\]",
    "panic:",
    "Exception:",
]
max_lines = 50
```

## Testing Your Filter

```bash
# Validate filter syntax
tokman filter validate ~/.config/tokman/filters/my-tool.toml

# Test filter against sample input
echo "Build starting...
[info] Compiling module A
[info] Compiling module B
[DEBUG] Loading cache
Build succeeded in 12.3s
2 warnings found" | tokman filter test my_tool_build

# Expected output:
# Build succeeded in 12.3s
# 2 warnings found
```

## Filter Precedence

1. User filters (`~/.config/tokman/filters/`) override built-in
2. More specific match patterns take priority
3. First matching filter wins (order by specificity)

## Tips

1. **Start broad, refine** - Begin with simple strip patterns, add more as needed
2. **Keep failures** - Always preserve error/failure output
3. **Test first** - Use `tokman filter test` before deploying
4. **Version control** - Keep filters in your dotfiles repo
5. **Share filters** - Contribute useful filters back to TokMan!
