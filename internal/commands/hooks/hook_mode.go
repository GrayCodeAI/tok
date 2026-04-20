package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Go-native replacement for hooks/tok-mode-{activate,config,tracker}.js.
// Sharing the Node.js flag file format means the two implementations
// interoperate during the migration period.

var validModes = map[string]bool{
	"off": true, "lite": true, "full": true, "ultra": true,
	"wenyan-lite": true, "wenyan": true, "wenyan-full": true, "wenyan-ultra": true,
	"commit": true, "review": true, "compress": true,
}

const maxFlagBytes = 64
const maxStdinBytes = 256 * 1024

var hookModeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Manage tok mode activation (Go-native; replaces Node.js hooks)",
	Long: `Go-native implementation of the SessionStart and UserPromptSubmit
hooks currently provided by hooks/tok-mode-*.js. Drop-in replacement —
writes/reads the same flag file format.

Subcommands:
  activate   SessionStart hook: emit current mode + skill rules
  track      UserPromptSubmit hook: inspect prompt, update mode flag
  status     print current mode (non-hook use)
  set <mode> write flag file manually`,
}

var hookModeActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "SessionStart hook body — emit mode rules and write flag",
	RunE:  runHookModeActivate,
}

var hookModeTrackCmd = &cobra.Command{
	Use:   "track",
	Short: "UserPromptSubmit hook body — parse prompt, update flag",
	RunE:  runHookModeTrack,
}

var hookModeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the active mode (or empty if none)",
	RunE:  runHookModeStatus,
}

var hookModeSetCmd = &cobra.Command{
	Use:   "set <mode>",
	Short: "Write the mode flag explicitly",
	Args:  cobra.ExactArgs(1),
	RunE:  runHookModeSet,
}

func init() {
	hookModeCmd.AddCommand(hookModeActivateCmd, hookModeTrackCmd, hookModeStatusCmd, hookModeSetCmd)
	hookCmd.AddCommand(hookModeCmd)
}

func claudeConfigDir() string {
	if v := os.Getenv("CLAUDE_CONFIG_DIR"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

func flagPath() string {
	return filepath.Join(claudeConfigDir(), ".tok-active")
}

func resolveDefaultMode() string {
	if v := os.Getenv("TOK_DEFAULT_MODE"); v != "" {
		v = strings.ToLower(v)
		if validModes[v] {
			return v
		}
	}
	// Config file lookup (mirrors tok-mode-config.js)
	configPaths := []string{}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		configPaths = append(configPaths, filepath.Join(xdg, "tok", "config.json"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		configPaths = append(configPaths, filepath.Join(home, ".config", "tok", "config.json"))
	}
	for _, p := range configPaths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var cfg struct {
			DefaultMode string `json:"defaultMode"`
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		m := strings.ToLower(cfg.DefaultMode)
		if validModes[m] {
			return m
		}
	}
	return "full"
}

// writeFlag writes the mode flag symlink-safely: refuses symlinks at the
// target and at the immediate parent, writes via temp + rename with 0600.
func writeFlag(mode string) error {
	if !validModes[mode] {
		return fmt.Errorf("invalid mode: %s", mode)
	}
	fp := flagPath()
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	// Refuse if parent dir is a symlink.
	if st, err := os.Lstat(dir); err == nil && st.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing: %s is a symlink", dir)
	}
	// Refuse if target exists as symlink.
	if st, err := os.Lstat(fp); err == nil && st.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing: %s is a symlink", fp)
	}
	tmp := fp + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(mode); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, fp)
}

// readFlag is size-capped + whitelist-validated.
func readFlag() string {
	fp := flagPath()
	st, err := os.Lstat(fp)
	if err != nil || st.Mode()&os.ModeSymlink != 0 || !st.Mode().IsRegular() || st.Size() > maxFlagBytes {
		return ""
	}
	data, err := os.ReadFile(fp)
	if err != nil {
		return ""
	}
	s := strings.ToLower(strings.TrimSpace(string(data)))
	if !validModes[s] {
		return ""
	}
	return s
}

func clearFlag() {
	_ = os.Remove(flagPath())
}

func runHookModeActivate(cmd *cobra.Command, args []string) error {
	mode := resolveDefaultMode()
	if mode == "off" {
		clearFlag()
		fmt.Print("OK")
		return nil
	}
	if err := writeFlag(mode); err != nil {
		// Best-effort; don't fail the hook.
		_ = err
	}
	label := mode
	if mode == "wenyan" {
		label = "wenyan-full"
	}
	// Emit concise fallback rules. A richer implementation reads SKILL.md and
	// filters the intensity table to the active level, matching the JS hook.
	fmt.Printf("TOK MODE ACTIVE — level: %s\n\n", label)
	fmt.Println("Respond terse like smart tok. Drop articles, filler, pleasantries, hedging.")
	fmt.Println("Fragments OK. Technical terms exact. Code blocks unchanged. Errors quoted exact.")
	fmt.Println("Persist every response until 'stop tok' or 'normal mode'.")
	return nil
}

func runHookModeTrack(cmd *cobra.Command, args []string) error {
	buf, err := io.ReadAll(io.LimitReader(os.Stdin, maxStdinBytes))
	if err != nil || len(buf) == 0 {
		return nil
	}
	var payload struct {
		Prompt string `json:"prompt"`
	}
	_ = json.Unmarshal(buf, &payload)
	prompt := strings.ToLower(strings.TrimSpace(payload.Prompt))

	// Activation triggers
	activate := containsAll(prompt, []string{"tok"}) && containsAny(prompt, []string{"activate", "enable", "turn on", "start", "talk like"})
	deactivate := containsAll(prompt, []string{"tok"}) && containsAny(prompt, []string{"stop", "disable", "turn off", "deactivate"})
	if strings.Contains(prompt, "normal mode") {
		deactivate = true
	}

	if deactivate {
		clearFlag()
		return nil
	}
	if activate {
		mode := resolveDefaultMode()
		if mode != "off" {
			_ = writeFlag(mode)
		}
	}

	// /tok slash commands
	if strings.HasPrefix(prompt, "/tok") {
		parts := strings.Fields(prompt)
		cmd := parts[0]
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}
		var mode string
		switch cmd {
		case "/tok-commit":
			mode = "commit"
		case "/tok-review":
			mode = "review"
		case "/tok-compress", "/tok:tok-compress":
			mode = "compress"
		case "/tok", "/tok:tok":
			switch arg {
			case "lite", "ultra", "wenyan-lite", "wenyan-full", "wenyan-ultra":
				mode = arg
			case "wenyan":
				mode = "wenyan"
			case "off":
				clearFlag()
			default:
				mode = resolveDefaultMode()
			}
		}
		if mode != "" && mode != "off" {
			_ = writeFlag(mode)
		}
	}

	// Per-turn reinforcement (matches JS tracker)
	active := readFlag()
	independent := map[string]bool{"commit": true, "review": true, "compress": true}
	if active != "" && !independent[active] {
		msg := fmt.Sprintf("TOK MODE ACTIVE (%s). Drop articles/filler/pleasantries/hedging. Fragments OK. Code/commits/security: write normal.", active)
		out := map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":     "UserPromptSubmit",
				"additionalContext": msg,
			},
		}
		enc := json.NewEncoder(os.Stdout)
		_ = enc.Encode(out)
	}
	return nil
}

func runHookModeStatus(cmd *cobra.Command, args []string) error {
	fmt.Println(readFlag())
	return nil
}

func runHookModeSet(cmd *cobra.Command, args []string) error {
	return writeFlag(strings.ToLower(args[0]))
}

func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
