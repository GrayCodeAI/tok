---
name: tok-help
description: >
  Quick-reference card for all tok modes, skills, and commands.
  One-shot display, not a persistent mode. Trigger: /tok-help,
  "tok help", "what tok commands", "how do I use tok".
---

# Tok Help

Display this reference card when invoked. One-shot — do NOT change mode, write flag files, or persist anything. Output in tok style.

## Modes

| Mode | Trigger | What change |
|------|---------|-------------|
| **Lite** | `/tok lite` | Drop filler. Keep sentence structure. |
| **Full** | `/tok` | Drop articles, filler, pleasantries, hedging. Fragments OK. Default. |
| **Ultra** | `/tok ultra` | Extreme compression. Bare fragments. Tables over prose. |
| **Wenyan-Lite** | `/tok wenyan-lite` | Classical Chinese style, light compression. |
| **Wenyan-Full** | `/tok wenyan` | Full 文言文. Maximum classical terseness. |
| **Wenyan-Ultra** | `/tok wenyan-ultra` | Extreme. Ancient scholar on a budget. |

Mode stick until changed or session end.

## Skills

| Skill | Trigger | What it do |
|-------|---------|-----------|
| **tok-commit** | `/tok-commit` | Terse commit messages. Conventional Commits. ≤50 char subject. |
| **tok-review** | `/tok-review` | One-line PR comments: `L42: bug: user null. Add guard.` |
| **tok-compress** | `/tok:compress <file>` | Compress .md files to tok prose. Saves ~46% input tokens. |
| **tok-help** | `/tok-help` | This card. |

## Deactivate

Say "stop tok" or "normal mode". Resume anytime with `/tok`.

## Configure Default Mode

Default mode = `full`. Change it:

**Environment variable** (highest priority):
```bash
export TOK_DEFAULT_MODE=ultra
```

**Config file** (`~/.config/tok/config.json`):
```json
{ "defaultMode": "lite" }
```

Set `"off"` to disable auto-activation on session start. User can still activate manually with `/tok`.

Resolution: env var > config file > `full`.

## More

Full docs: https://github.com/JuliusBrussee/tok
