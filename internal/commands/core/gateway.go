package core

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "AI Gateway features: kill switches, quotas, model routing",
	Long: `AI Gateway provides kill switches, quotas, model aliasing,
and fallback chains for team AI usage.

Examples:
  tokman gateway status
  tokman gateway set-quota gpt-4 10000
  tokman gateway alias gpt-4 gpt-4o-mini`,
	RunE: runGateway,
}

var (
	gwAction     string
	gwModel      string
	gwAlias      string
	gwQuota      int
	gwKillSwitch bool
)

// GatewayConfig holds gateway configuration.
type GatewayConfig struct {
	KillSwitches map[string]bool
	Quotas       map[string]int
	ModelAliases map[string]string
	Fallbacks    map[string][]string
}

// DefaultGatewayConfig returns default gateway configuration.
func DefaultGatewayConfig() GatewayConfig {
	return GatewayConfig{
		KillSwitches: make(map[string]bool),
		Quotas:       make(map[string]int),
		ModelAliases: make(map[string]string),
		Fallbacks:    make(map[string][]string),
	}
}

func init() {
	registry.Add(func() { registry.Register(gatewayCmd) })
	gatewayCmd.Flags().StringVar(&gwAction, "action", "status", "Action: status, set-quota, alias, kill-switch, fallback")
	gatewayCmd.Flags().StringVar(&gwModel, "model", "", "Model name")
	gatewayCmd.Flags().StringVar(&gwAlias, "alias", "", "Alias target")
	gatewayCmd.Flags().IntVar(&gwQuota, "quota", 0, "Token quota")
	gatewayCmd.Flags().BoolVar(&gwKillSwitch, "kill-switch", false, "Enable/disable kill switch")
}

func runGateway(cmd *cobra.Command, args []string) error {
	config := DefaultGatewayConfig()

	switch gwAction {
	case "status":
		fmt.Println("Gateway Status:")
		fmt.Println("  Kill Switches: none active")
		fmt.Println("  Quotas: none set")
		fmt.Println("  Aliases: none configured")
		fmt.Println("  Fallbacks: none configured")

	case "set-quota":
		if gwModel == "" || gwQuota == 0 {
			return fmt.Errorf("--model and --quota required")
		}
		config.Quotas[gwModel] = gwQuota
		fmt.Printf("Quota set: %s = %d tokens\n", gwModel, gwQuota)

	case "alias":
		if gwModel == "" || gwAlias == "" {
			return fmt.Errorf("--model and --alias required")
		}
		config.ModelAliases[gwModel] = gwAlias
		fmt.Printf("Alias set: %s → %s\n", gwModel, gwAlias)

	case "kill-switch":
		if gwModel == "" {
			return fmt.Errorf("--model required")
		}
		config.KillSwitches[gwModel] = gwKillSwitch
		status := "disabled"
		if gwKillSwitch {
			status = "enabled"
		}
		fmt.Printf("Kill switch %s for %s\n", status, gwModel)

	case "fallback":
		if gwModel == "" || gwAlias == "" {
			return fmt.Errorf("--model and --alias required")
		}
		config.Fallbacks[gwModel] = append(config.Fallbacks[gwModel], gwAlias)
		fmt.Printf("Fallback added: %s → %s\n", gwModel, gwAlias)

	default:
		return fmt.Errorf("unknown action: %s", gwAction)
	}

	return nil
}

// CheckQuota checks if a model has exceeded its quota.
func CheckQuota(quotas map[string]int, model string, used int) bool {
	limit, ok := quotas[model]
	if !ok {
		return false
	}
	return used >= limit
}

// ResolveModel resolves a model through aliases and fallbacks.
func ResolveModel(aliases map[string]string, fallbacks map[string][]string, model string) string {
	if alias, ok := aliases[model]; ok {
		return alias
	}
	if fb, ok := fallbacks[model]; ok && len(fb) > 0 {
		return fb[0]
	}
	return model
}

// CheckKillSwitch checks if a model is kill-switched.
func CheckKillSwitch(switches map[string]bool, model string) bool {
	return switches[model]
}

// FormatGatewayStatus returns a human-readable status string.
func FormatGatewayStatus(config GatewayConfig) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Kill Switches: %d", len(config.KillSwitches)))
	parts = append(parts, fmt.Sprintf("Quotas: %d", len(config.Quotas)))
	parts = append(parts, fmt.Sprintf("Aliases: %d", len(config.ModelAliases)))
	parts = append(parts, fmt.Sprintf("Fallbacks: %d", len(config.Fallbacks)))
	return strings.Join(parts, ", ")
}
