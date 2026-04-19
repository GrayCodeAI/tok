# Tok Hooks

Optional shell helpers for users who want tok mode/status surfaced in shell prompts.

## Files

- `tok-statusline.sh` prints `[TOK]` badge based on active mode file.
- `tok-statusline.ps1` prints status badge for PowerShell.
- `install.sh` adds an optional shell snippet to `~/.zshrc` and `~/.bashrc`.
- `uninstall.sh` removes the snippet.
- `install.ps1` / `uninstall.ps1` do the same for PowerShell profile.

## Install

```bash
bash hooks/install.sh
```

```powershell
powershell -ExecutionPolicy Bypass -File hooks\install.ps1
```

## Uninstall

```bash
bash hooks/uninstall.sh
```

```powershell
powershell -ExecutionPolicy Bypass -File hooks\uninstall.ps1
```
