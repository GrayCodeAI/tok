package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"

	"github.com/GrayCodeAI/tok/internal/config"
)

// hasKeybindings returns true if any keybinding override is set.
func hasKeybindings(cfg config.KeybindingsConfig) bool {
	return cfg.Quit != "" || cfg.NextSection != "" || cfg.PrevSection != "" ||
		cfg.HistoryBack != "" || cfg.HistoryForward != "" || cfg.Refresh != "" ||
		cfg.Up != "" || cfg.Down != "" || cfg.PageUp != "" || cfg.PageDown != "" ||
		cfg.Top != "" || cfg.Bottom != "" || cfg.Enter != "" || cfg.Back != "" ||
		cfg.Yank != "" || cfg.Export != "" || cfg.Palette != "" || cfg.Search != "" ||
		cfg.Help != ""
}

// LoadKeyMap merges user keybindings from config with defaults.
// User bindings override defaults when specified.
func LoadKeyMap(cfg config.KeybindingsConfig) (KeyMap, error) {
	km := DefaultKeyMap()
	var errors []string

	if cfg.Quit != "" {
		if err := validateKey("quit", cfg.Quit); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Quit = key.NewBinding(
				key.WithKeys(cfg.Quit),
				key.WithHelp(cfg.Quit, "quit"),
			)
		}
	}
	if cfg.NextSection != "" {
		if err := validateKey("next_section", cfg.NextSection); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.NextSection = key.NewBinding(
				key.WithKeys(cfg.NextSection),
				key.WithHelp(cfg.NextSection, "next section"),
			)
		}
	}
	if cfg.PrevSection != "" {
		if err := validateKey("prev_section", cfg.PrevSection); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.PrevSection = key.NewBinding(
				key.WithKeys(cfg.PrevSection),
				key.WithHelp(cfg.PrevSection, "prev section"),
			)
		}
	}
	if cfg.HistoryBack != "" {
		if err := validateKey("history_back", cfg.HistoryBack); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.HistoryBack = key.NewBinding(
				key.WithKeys(cfg.HistoryBack),
				key.WithHelp(cfg.HistoryBack, "history back"),
			)
		}
	}
	if cfg.HistoryForward != "" {
		if err := validateKey("history_forward", cfg.HistoryForward); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.HistoryForward = key.NewBinding(
				key.WithKeys(cfg.HistoryForward),
				key.WithHelp(cfg.HistoryForward, "history forward"),
			)
		}
	}
	if cfg.Refresh != "" {
		if err := validateKey("refresh", cfg.Refresh); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Refresh = key.NewBinding(
				key.WithKeys(cfg.Refresh),
				key.WithHelp(cfg.Refresh, "refresh"),
			)
		}
	}
	if cfg.Up != "" {
		if err := validateKey("up", cfg.Up); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Up = key.NewBinding(
				key.WithKeys(cfg.Up),
				key.WithHelp(cfg.Up, "cursor up"),
			)
		}
	}
	if cfg.Down != "" {
		if err := validateKey("down", cfg.Down); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Down = key.NewBinding(
				key.WithKeys(cfg.Down),
				key.WithHelp(cfg.Down, "cursor down"),
			)
		}
	}
	if cfg.PageUp != "" {
		if err := validateKey("page_up", cfg.PageUp); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.PageUp = key.NewBinding(
				key.WithKeys(cfg.PageUp),
				key.WithHelp(cfg.PageUp, "page up"),
			)
		}
	}
	if cfg.PageDown != "" {
		if err := validateKey("page_down", cfg.PageDown); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.PageDn = key.NewBinding(
				key.WithKeys(cfg.PageDown),
				key.WithHelp(cfg.PageDown, "page down"),
			)
		}
	}
	if cfg.Top != "" {
		if err := validateKey("top", cfg.Top); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Top = key.NewBinding(
				key.WithKeys(cfg.Top),
				key.WithHelp(cfg.Top, "jump to top"),
			)
		}
	}
	if cfg.Bottom != "" {
		if err := validateKey("bottom", cfg.Bottom); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Bottom = key.NewBinding(
				key.WithKeys(cfg.Bottom),
				key.WithHelp(cfg.Bottom, "jump to bottom"),
			)
		}
	}
	if cfg.Enter != "" {
		if err := validateKey("enter", cfg.Enter); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Enter = key.NewBinding(
				key.WithKeys(cfg.Enter),
				key.WithHelp(cfg.Enter, "drill into row"),
			)
		}
	}
	if cfg.Back != "" {
		if err := validateKey("back", cfg.Back); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Back = key.NewBinding(
				key.WithKeys(cfg.Back),
				key.WithHelp(cfg.Back, "back"),
			)
		}
	}
	if cfg.Yank != "" {
		if err := validateKey("yank", cfg.Yank); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Yank = key.NewBinding(
				key.WithKeys(cfg.Yank),
				key.WithHelp(cfg.Yank, "yank row"),
			)
		}
	}
	if cfg.Export != "" {
		if err := validateKey("export", cfg.Export); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Export = key.NewBinding(
				key.WithKeys(cfg.Export),
				key.WithHelp(cfg.Export, "export view"),
			)
		}
	}
	if cfg.Palette != "" {
		if err := validateKey("palette", cfg.Palette); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Palette = key.NewBinding(
				key.WithKeys(cfg.Palette),
				key.WithHelp(cfg.Palette, "command palette"),
			)
		}
	}
	if cfg.Search != "" {
		if err := validateKey("search", cfg.Search); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Search = key.NewBinding(
				key.WithKeys(cfg.Search),
				key.WithHelp(cfg.Search, "search"),
			)
		}
	}
	if cfg.Help != "" {
		if err := validateKey("help", cfg.Help); err != nil {
			errors = append(errors, err.Error())
		} else {
			km.Help = key.NewBinding(
				key.WithKeys(cfg.Help),
				key.WithHelp(cfg.Help, "help"),
			)
		}
	}

	if len(errors) > 0 {
		return km, fmt.Errorf("keybinding errors: %s", strings.Join(errors, "; "))
	}
	return km, nil
}

// validateKey checks if a key string is valid for bubbletea.
// Rejects empty strings and single-character modifiers.
func validateKey(name, key string) error {
	if key == "" {
		return fmt.Errorf("%s: empty key not allowed", name)
	}

	// Reject single-character modifier-looking keys
	if key == "ctrl" || key == "alt" || key == "shift" || key == "super" {
		return fmt.Errorf("%s: '%s' is a modifier, not a key", name, key)
	}

	// Check for duplicate commas which would create empty keys
	if strings.Contains(key, ",,") {
		return fmt.Errorf("%s: invalid key sequence with empty key", name)
	}

	return nil
}
