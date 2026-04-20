# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.29.x  | ✅ |
| < 0.29  | ❌ |

## Reporting a Vulnerability

We take security seriously. If you discover a vulnerability:

1. **Do NOT** open a public issue
2. Email us at [security@graycode.ai](mailto:security@graycode.ai) or use [GitHub Security Advisories](https://github.com/GrayCodeAI/tok/security/advisories)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will respond within **48 hours** and aim to release a fix within **14 days**.

## Security Considerations

### Input Handling

- tok processes untrusted input from terminal output and user input
- All input is validated for size (50MB limit) and encoding
- Shell scripts executed by hook commands are validated for path traversal and permissions

### Data Storage

- Tracking data is stored locally in SQLite (`~/.local/share/tok/tracking.db`)
- No data is transmitted externally
- Telemetry is anonymous and opt-out

### Dependencies

- We monitor dependencies via Dependabot (`.github/dependabot.yml`)
- All dependencies are pinned in `go.sum`
- SBOM is generated for each release (CycloneDX format)

### Best Practices

- All secrets should be environment variables, never hardcoded
- No credentials in source code, commits, or logs
- Use `go vet` and `golangci-lint` for static analysis
- Race detector enabled in CI (`make test-race`)

## Threat Model

tok sits between your AI coding agent and the shell. It installs
per-agent hook scripts (`tok-rewrite.sh`) that intercept every bash
command the agent would run. Because those hooks execute before the
agent's permission prompts, any unauthorized modification to a hook is
a command-injection vector against the agent's entire session.

### Assets

- **Hook scripts**: `~/.claude/hooks/tok-rewrite.sh` and equivalents
  under `~/.cursor/`, `~/.gemini/`, `~/.qwen/`, `~/.config/opencode/`,
  etc. Execute with the user's shell privileges on every agent tool call.
- **Agent settings files** patched by `tok init` (Claude `settings.json`,
  Cursor `hooks.json`, Gemini `settings.json`). Losing integrity on these
  means the hook wiring itself can be redirected.
- **Tracking database** at `~/.local/share/tok/tracking.db`. Contains
  command history and token counts. Read-only locally; no network egress.

### Adversary capabilities we defend against

- **Accidental edits**: user or another tool rewrites a hook script or
  settings file. Mitigation: SHA-256 baseline per hook, `tok doctor
  --security` audit, runtime integrity check in the Claude hook.
- **Supply-chain drift**: a dependency upgrade changes a hook's generated
  content unexpectedly. Mitigation: hook version marker
  (`# tok-hook-version: N`) + explicit outdated status; idempotent
  re-install re-records the baseline.

### Adversary capabilities we do NOT defend against

- **Local attacker with write access to `~/.claude/hooks/`**: they can
  replace both the hook and its baseline hash file, defeating integrity.
  tok does not enforce filesystem permissions beyond making the hash
  file read-only (0444), which is a speed bump, not a boundary.
- **Root or equivalent**: full system compromise is out of scope.
- **Compromised tok binary itself**: verify your install via
  `shasum -a 256` against the release checksum.

### Operator runbook

- Before a pairing session or demo: `tok doctor --security`. Exits
  non-zero if any hook is tampered or has no baseline.
- After a suspected incident: delete the affected hook + hash file,
  then `tok init --<agent>` to reinstall from a known-good source.
- To audit across all installed agents in one go: the security-only
  output lists each hook path, its expected/actual SHA prefix, and
  recommended remediation per status.

### Runtime integrity gates

Today, only the Claude Code hook calls `tok hook claude`, which in
turn invokes `integrity.RuntimeCheck` and fails closed on tamper. The
other 19 wired agents' hooks do not yet invoke a runtime gate; they
rely on the `tok doctor --security` audit running before a session.
Wiring RuntimeCheck into every agent's hook template is planned work
tracked against the hook infrastructure.
