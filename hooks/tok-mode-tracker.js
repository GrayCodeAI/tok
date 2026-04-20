#!/usr/bin/env node
// tok — UserPromptSubmit hook to track which tok mode is active
// Inspects user input for /tok commands and writes mode to flag file

const fs = require('fs');
const path = require('path');
const os = require('os');
const { getDefaultMode, safeWriteFlag, readFlag } = require('./tok-mode-config');

const claudeDir = process.env.CLAUDE_CONFIG_DIR || path.join(os.homedir(), '.claude');
const flagPath = path.join(claudeDir, '.tok-active');

// Hard cap on stdin bytes — Claude Code passes the user prompt as JSON, but a
// malicious caller could pipe unbounded data to balloon memory before parse.
const MAX_STDIN_BYTES = 256 * 1024;
let input = '';
let overflow = false;
process.stdin.on('data', chunk => {
  if (overflow) return;
  if (input.length + chunk.length > MAX_STDIN_BYTES) {
    overflow = true;
    try { process.stdin.destroy(); } catch (e) {}
    return;
  }
  input += chunk;
});
process.stdin.on('end', () => {
  if (overflow) return;
  try {
    const data = JSON.parse(input);
    const prompt = (data.prompt || '').trim().toLowerCase();

    // Natural language activation (e.g. "activate tok", "turn on tok mode",
    // "talk like tok"). README tells users they can say these, but the hook
    // only matched /tok commands — flag file and statusline stayed out of sync.
    if (/\b(activate|enable|turn on|start|talk like)\b.*\btok\b/i.test(prompt) ||
        /\btok\b.*\b(mode|activate|enable|turn on|start)\b/i.test(prompt)) {
      if (!/\b(stop|disable|turn off|deactivate)\b/i.test(prompt)) {
        const mode = getDefaultMode();
        if (mode !== 'off') {
          safeWriteFlag(flagPath, mode);
        }
      }
    }

    // Match /tok commands
    if (prompt.startsWith('/tok')) {
      const parts = prompt.split(/\s+/);
      const cmd = parts[0]; // /tok, /tok-commit, /tok-review, etc.
      const arg = parts[1] || '';

      let mode = null;

      if (cmd === '/tok-commit') {
        mode = 'commit';
      } else if (cmd === '/tok-review') {
        mode = 'review';
      } else if (cmd === '/tok-compress' || cmd === '/tok:tok-compress') {
        mode = 'compress';
      } else if (cmd === '/tok' || cmd === '/tok:tok') {
        if (arg === 'lite') mode = 'lite';
        else if (arg === 'ultra') mode = 'ultra';
        else if (arg === 'wenyan-lite') mode = 'wenyan-lite';
        else if (arg === 'wenyan' || arg === 'wenyan-full') mode = 'wenyan';
        else if (arg === 'wenyan-ultra') mode = 'wenyan-ultra';
        else mode = getDefaultMode();
      }

      if (mode && mode !== 'off') {
        safeWriteFlag(flagPath, mode);
      } else if (mode === 'off') {
        try { fs.unlinkSync(flagPath); } catch (e) {}
      }
    }

    // Detect deactivation — natural language and slash commands
    if (/\b(stop|disable|deactivate|turn off)\b.*\btok\b/i.test(prompt) ||
        /\btok\b.*\b(stop|disable|deactivate|turn off)\b/i.test(prompt) ||
        /\bnormal mode\b/i.test(prompt)) {
      try { fs.unlinkSync(flagPath); } catch (e) {}
    }

    // Per-turn reinforcement: emit a structured reminder when tok is active.
    // The SessionStart hook injects the full ruleset once, but models lose it
    // when other plugins inject competing style instructions every turn.
    // This keeps tok visible in the model's attention on every user message.
    //
    // Skip independent modes (commit, review, compress) — they have their own
    // skill behavior and the base tok rules would conflict.
    // readFlag enforces symlink-safe read + size cap + VALID_MODES whitelist.
    // If the flag is missing, corrupted, oversized, or a symlink pointing at
    // something like ~/.ssh/id_rsa, readFlag returns null and we emit nothing
    // — never inject untrusted bytes into model context.
    const INDEPENDENT_MODES = new Set(['commit', 'review', 'compress']);
    const activeMode = readFlag(flagPath);
    if (activeMode && !INDEPENDENT_MODES.has(activeMode)) {
      process.stdout.write(JSON.stringify({
        hookSpecificOutput: {
          hookEventName: "UserPromptSubmit",
          additionalContext: "TOK MODE ACTIVE (" + activeMode + "). " +
            "Drop articles/filler/pleasantries/hedging. Fragments OK. " +
            "Code/commits/security: write normal."
        }
      }));
    }
  } catch (e) {
    // Silent fail
  }
});
