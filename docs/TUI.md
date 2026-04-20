# tok TUI

Interactive terminal dashboard for tok — `tok tui` launches a Bubble Tea
app that surfaces every piece of data tok records (savings, costs,
sessions, pipeline layers, streaks, logs) through keyboard-driven
navigation, with first-class drill-down, search, and export.

---

## Launch

```bash
tok tui                        # default 30-day window, dark theme
tok tui --days 7               # narrower window
tok tui --agent claude         # filter to one agent
tok tui --theme light          # swap theme at startup
tok tui --refresh 5s           # faster live refresh
```

All flags:

| Flag | Default | Meaning |
|------|---------|---------|
| `--refresh` | `20s` | how often to re-read the workspace snapshot |
| `--days` | `30` | active-window size for trends/totals |
| `--project` | (all) | limit to one project path |
| `--agent` | (all) | limit to one agent (claude, copilot, ...) |
| `--provider` | (all) | limit to one provider (anthropic, openai, ...) |
| `--model` | (all) | limit to one model |
| `--session` | (all) | limit to one session ID |
| `--theme` | `dark` | `dark` · `light` · `high-contrast` · `colorblind` |

`tok tui` refuses to start unless stdin and stdout are both TTYs — it
won't corrupt a pipe with escape sequences.

---

## Sections

| # | Name | Purpose |
|---|------|---------|
| 1 | **Home** | overview cockpit: saved tokens, cost, reduction, streaks, leaderboards |
| 2 | **Today** | newest daily bucket with deltas vs yesterday + 7-day sparkline |
| 3 | **Trends** | multi-row Braille line charts for saved / reduction / commands |
| 4 | **Providers** | rank by provider; drill into per-model breakdown |
| 5 | **Models** | rank by model; drill into provider partners |
| 6 | **Agents** | rank by agent; drill into recent sessions for the agent |
| 7 | **Sessions** | list every recent session; drill into identity/activity/snapshots |
| 8 | **Commands** | top-performing vs weakest commands (`t`/`w` toggle) |
| 9 | **Pipeline** | per-layer contribution bars + sortable table |
| 10 | **Rewards** | streak calendar, points, level, badges |
| 11 | **Logs** | live in-memory log ring with level filter + search |
| 12 | **Config** | hook state (toggle with `t`), paths, data-quality, active filters |

---

## Keybindings

All bindings are declared in `internal/tui/keys.go`. The `?` overlay
inside the app auto-generates from that registry, so it's always
current — the table below is a convenience snapshot.

### Global

| Key | Action |
|-----|--------|
| `1`–`9` | jump to section N |
| `tab` / `→` / `l` | next section |
| `shift+tab` / `←` / `h` | prev section |
| `r` | refresh snapshot |
| `:` | open command palette |
| `/` | open in-pane search |
| `e` | export current section to `~/.tok/exports/<section>-<ts>.json` |
| `?` | toggle help overlay |
| `esc` | close open overlay |
| `q` / `ctrl+c` | quit (cancels in-flight queries, closes DB) |

### Cursor (any section with a table)

| Key | Action |
|-----|--------|
| `j` / `↓` | cursor down |
| `k` / `↑` | cursor up |
| `g` / `home` | top of list |
| `G` / `end` | bottom of list |
| `pgup` / `ctrl+b` | page up |
| `pgdn` / `ctrl+f` | page down |
| `enter` | drill into selected row |
| `backspace` | exit drill, return to list |
| `y` | yank selected row as TSV to clipboard (OSC-52) |

### Section-local

| Section | Key | Action |
|---------|-----|--------|
| Trends | `d` / `w` | daily / weekly granularity |
| Commands | `t` / `w` | top / weak list |
| Logs | `d` / `i` / `w` / `e` | debug+ / info+ / warn+ / error only |
| Logs | `c` | clear log ring (prompts for confirmation) |
| Config | `t` | toggle tok shell hook |

---

## Command palette

Press `:` anywhere to open. Start typing to fuzzy-match across every
registered action and every section name.

Built-in actions (extend by registering in `DefaultActionRegistry`):

| ID | What it does | Confirm? |
|----|--------------|----------|
| `view.refresh` | reload workspace snapshot | no |
| `section.jump <n>` | jump to section by 1-based index | no |
| `theme.set <name>` | set theme by name | no |
| `theme.cycle` | advance to next bundled theme | no |
| `toast.info [msg]` | emit an info toast (diagnostic) | no |
| `hooks.toggle` | flip the global tok shell hook | no (reversible) |
| `logs.clear` | drop every in-memory log event | **yes** |

---

## Clipboard (OSC-52)

`y` in any table section copies the selected row as tab-separated
values via the OSC-52 escape sequence. Works over ssh in any terminal
that honors OSC-52 (kitty, iTerm2, wezterm, alacritty, Windows
Terminal; xterm with allowClipboardOps=true). No native clipboard
daemon required.

If the terminal silently drops the sequence, the toast still confirms —
inspect your terminal's clipboard-passthrough docs if `y` doesn't
populate the system clipboard.

---

## Theme switching

Four bundled themes:

- **dark** — default; truecolor accents, panel backgrounds
- **light** — muted backgrounds for bright terminals
- **high-contrast** — pure black/white + saturated accents; aim for
  WCAG AAA on bold text
- **colorblind** — [Okabe-Ito](https://jfly.uni-koeln.de/color/) 8-color
  palette, safe for protan/deutan/tritan viewers

Change at runtime with `:theme.set <name>` or cycle with
`:theme.cycle`.

---

## Unicode / ASCII fallback

When `$LC_ALL` / `$LC_CTYPE` / `$LANG` don't advertise UTF-8, the TUI
swaps Unicode glyphs for ASCII substitutes automatically:

| Context | Unicode | ASCII |
|---------|---------|-------|
| block sparkline | `▁▂▃▄▅▆▇█` | `.-=#` |
| Braille line chart | Braille U+2800..U+28FF | `*` markers on a grid |
| bar fill | `█` / `░` | `#` / `-` |
| streak calendar | `░ ▒ ▓ █` | `. : + #` |

The rest of the layout is always UTF-8-agnostic (box-drawing falls
through to lipgloss, which is already ASCII-safe at compile time).

---

## Adding a section

1. Create `internal/tui/view_<name>.go` with a struct implementing
   [`SectionRenderer`](../internal/tui/sections.go):

   ```go
   func (s *mySection) Name() string
   func (s *mySection) Short() string
   func (s *mySection) Init(SectionContext) tea.Cmd
   func (s *mySection) KeyBindings() []key.Binding
   func (s *mySection) Update(SectionContext, tea.Msg) (SectionRenderer, tea.Cmd)
   func (s *mySection) View(SectionContext) string
   ```

2. Register it in [`defaultSections()`](../internal/tui/sections.go)
   replacing the placeholder slot.

3. Read data from `ctx.Data` (the shared `WorkspaceDashboardSnapshot`)
   and `ctx.Logs` (the in-memory slog ring).

4. For breakdown-style views, embed a `*Table` and delegate nav to
   `handleTableNav`; for visual views, use `BrailleLineChart` /
   `LineChart` from `chart.go`.

5. To expose Export (`e`) on the section, also implement the
   [`ExportableTable`](../internal/tui/export.go) interface — three
   trivial getters wrapping the embedded `Table`.

6. If the section introduces a mutation, add it to
   `DefaultActionRegistry` in `actions.go`. Set `Confirm: true` when
   the operation is destructive; the root model will route it through
   the confirm modal automatically.

---

## Architecture

```
┌───────────────────────────────────────────────────────────────────┐
│ tea.Program                                                       │
│ ┌───────────────────────────────────────────────────────────────┐ │
│ │ root model (app.go)                                           │ │
│ │                                                               │ │
│ │ ┌──────────┐  ┌─────────────┐  ┌────────────┐  ┌────────────┐ │ │
│ │ │ keys.go  │  │ overlays    │  │ actions.go │  │ toasts     │ │ │
│ │ │ KeyMap   │  │ palette     │  │ registry   │  │ stack +    │ │ │
│ │ │ ShortHelp│  │ search      │  │ Run + Cmd  │  │ TTL tick   │ │ │
│ │ │ FullHelp │  │ confirm     │  │            │  │            │ │ │
│ │ └──────────┘  └─────────────┘  └────────────┘  └────────────┘ │ │
│ │                                                               │ │
│ │ ┌───────────────────────────────────────────────────────────┐ │ │
│ │ │ SectionRenderer dispatch (12 sections)                    │ │ │
│ │ │   view_home  view_today  view_trends  view_providers      │ │ │
│ │ │   view_models  view_agents  view_sessions  view_commands  │ │ │
│ │ │   view_pipeline  view_rewards  view_logs  view_config     │ │ │
│ │ └───────────────────────────────────────────────────────────┘ │ │
│ │                                                               │ │
│ │ ┌─────────────┐  ┌────────────┐  ┌──────────┐  ┌────────────┐ │ │
│ │ │ loader.go   │  │ logring.go │  │ chart.go │  │ table.go   │ │ │
│ │ │ Tracker +   │  │ slog ring  │  │ Braille  │  │ responsive │ │ │
│ │ │ SessionMgr  │  │ capture    │  │ + ASCII  │  │ + filter + │ │ │
│ │ │ long-lived  │  │ w/delegate │  │ sparklns │  │ sort + yank│ │ │
│ │ └─────────────┘  └────────────┘  └──────────┘  └────────────┘ │ │
│ └───────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────┘
                           │                       │
                           ▼                       ▼
                  SQLite: tracking.db       SQLite: sessions.db
                  tracking.Tracker          session.Manager
```

**Invariants** (documented in `internal/tui/doc.go`):

1. `internal/state.Global()` is populated once in the root command's
   `PersistentPreRunE`, before `runTUI` starts the tea.Program. The
   TUI never mutates it; no cobra subcommand runs during its lifetime.
2. The `snapshotLoader` owns both SQLite handles. On `q` the root
   model dispatches `shutdownCmd` which cancels the context, closes
   the loader, and restores the pre-TUI `slog.Default()` *before*
   sending `tea.Quit`.
3. Any stdlib code path that `fmt.Print`s to stdout inside the TUI's
   lifetime is a bug — the alt-screen will shred the frame. Route
   through `internal/output.Global()` (the compressor already does).

---

## Performance

Targets measured with `make benchmark-tui` on a 4-core EPYC; update
`internal/tui/bench_test.go` when the shape changes materially:

| Benchmark | Target | Observed |
|-----------|-------:|---------:|
| `BrailleLineChart_Wide` | < 250 µs/op | ~27 µs |
| `TableRender_1000Rows` | < 6 ms/op | ~1.0 ms |
| `ModelView_FullFrame` (16ms = 60fps budget) | < 16 ms/op | ~4.4 ms |
| `PaletteFuzzySearch` | < 50 µs/op | ~5 µs |

Run:

```bash
make benchmark-tui                              # writes artifacts/tui-bench.txt
```

---

## Testing

```bash
go test ./internal/tui/...                      # unit + golden
go test ./internal/tui/... -update -run TestGolden  # refresh goldens
```

Golden files live under `internal/tui/testdata/` and use
`lipgloss.SetColorProfile(termenv.Ascii)` so they're portable — no
host-specific truecolor escapes baked in. Run with `-update` after any
intentional layout or copy change, then inspect the diff before
committing.

Coverage of the TUI package crosses 60 tests (unit + golden snapshots)
plus the benchmark suite.
