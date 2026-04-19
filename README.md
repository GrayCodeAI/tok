# tok

Unified token optimization CLI for both input and output paths.

## Overview

tok is a unified CLI tool that combines:
- **Input path**: Compress human-written text before sending to AI
- **Output path**: Filter and compress terminal/tool output from AI

## Quick Start

```bash
# Build
cd tok
make build

# Test input compression
tok compress -mode ultra -input "Please implement a user authentication system"

# Test output filtering
tok git status

# Check unified status
tok doctor
```

## Commands

### Input Commands

```bash
tok input compress -mode ultra -input "text to compress"
tok input terse <mode>
tok input mode
tok input statusline
# ... and more
```

### Output Commands

```bash
tok output git status
tok output npm test
tok output cargo build
# ... and 100+ more
```

### Unified Commands

```bash
tok doctor      # Check both engines
tok status      # Show unified status
tok version     # Show version
tok both        # Run both input and output
```

## Architecture

- `tok input` - Native input compression engine (from tok)
- `tok output` - Output filtering engine (from tok)
- Single binary, no external dependencies

## Config

- Config: `~/.config/tok/config.toml`
- Data: `~/.local/share/tok/`

## License

MIT
