# tok for Gemini CLI

Auto-activates tok terse mode on session start.

## Auto-Activation

Add this to your `~/.gemini/settings.json` hooks:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "Always",
        "hooks": [
          {
            "type": "command",
            "command": "sh -c 'if command -v tok >/dev/null 2>&1; then mkdir -p ~/.config/tok && echo full > ~/.config/tok/.tok-active; fi'"
          }
        ]
      }
    ]
  }
}
```

## Modes

- `/tok lite` — Keep grammar, drop filler
- `/tok full` — Drop articles, fragments OK
- `/tok ultra` — Maximum compression, abbreviations
- `/tok off` — Normal mode

## Tagline

Same fix. 75% less word.
