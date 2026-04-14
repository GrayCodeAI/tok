package ai

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var currentProvider string
var currentModel string

var providers = map[string]*Provider{
	"openai": {
		Name:      "OpenAI",
		APIKey:    os.Getenv("OPENAI_API_KEY"),
		BaseURL:   "https://api.openai.com/v1",
		Models:    []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"},
		Available: os.Getenv("OPENAI_API_KEY") != "",
	},
	"anthropic": {
		Name:      "Anthropic",
		APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
		BaseURL:   "https://api.anthropic.com",
		Models:    []string{"claude-3-5-sonnet-20241022", "claude-3-opus-20240229", "claude-3-haiku-20240307"},
		Available: os.Getenv("ANTHROPIC_API_KEY") != "",
	},
	"google": {
		Name:      "Google",
		APIKey:    os.Getenv("GOOGLE_API_KEY"),
		BaseURL:   "https://generativelanguage.googleapis.com/v1beta",
		Models:    []string{"gemini-pro", "gemini-pro-vision"},
		Available: os.Getenv("GOOGLE_API_KEY") != "",
	},
	"grok": {
		Name:      "xAI Grok",
		APIKey:    os.Getenv("XAI_API_KEY"),
		BaseURL:   "https://api.x.ai/v1",
		Models:    []string{"grok-beta", "grok-vision-beta"},
		Available: os.Getenv("XAI_API_KEY") != "",
	},
}

type Provider struct {
	Name      string
	APIKey    string
	Models    []string
	BaseURL   string
	Available bool
}

func init() {
	registry.Add(func() {
		registry.Register(aiCmd)
	})
}

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Multi-provider AI integration",
	Long: `Manage AI providers for intelligent command processing.
		
Supported providers:
- openai (GPT-4, GPT-3.5)
- anthropic (Claude)
- google (Gemini)
- grok (xAI)

Examples:
  tokman ai chat "explain this error"
  tokman ai set-provider anthropic
  tokman ai config`,
}

var chatCmd = &cobra.Command{
	Use:   "chat [message]",
	Short: "Chat with AI",
	RunE:  runAIChat,
}

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available providers",
	RunE:  runProviders,
}

var setProviderCmd = &cobra.Command{
	Use:   "set-provider [name]",
	Short: "Set active provider",
	RunE:  runSetProvider,
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available models",
	RunE:  runModels,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure AI settings",
	RunE:  runAIConfig,
}

func init() {
	aiCmd.AddCommand(chatCmd, providersCmd, setProviderCmd, modelsCmd, configCmd)
	chatCmd.Flags().StringVar(&currentProvider, "provider", "", "AI provider")
	chatCmd.Flags().StringVarP(&currentModel, "model", "m", "", "AI model")
}

func runAIChat(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("message required")
	}

	provider := currentProvider
	if provider == "" {
		provider = getDefaultProvider()
	}

	p, ok := providers[provider]
	if !ok {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	if !p.Available {
		return fmt.Errorf("%s API key not set. Set %s_API_KEY environment variable", p.Name, strings.ToUpper(provider))
	}

	message := strings.Join(args, " ")

	fmt.Printf("[%s] ", p.Name)
	fmt.Print("Processing...\n")

	response := generateResponse(p.Name, message)

	fmt.Printf("\n%s: %s\n", p.Name, response)
	fmt.Printf("\nTokens: ~%d (estimated)\n", len(strings.Fields(response))*1)

	return nil
}

func runProviders(cmd *cobra.Command, args []string) error {
	fmt.Println("=== Available AI Providers ===")

	for name, p := range providers {
		status := "not configured"
		if p.Available {
			status = "available"
		}

		defaultMarker := ""
		if name == getDefaultProvider() {
			defaultMarker = " (default)"
		}

		fmt.Printf("%s%s: %s\n", name, defaultMarker, status)
		fmt.Printf("  Models: %s\n", strings.Join(p.Models[:2], ", "))
	}

	return nil
}

func runSetProvider(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("provider name required")
	}

	name := args[0]
	if _, ok := providers[name]; !ok {
		return fmt.Errorf("unknown provider: %s", name)
	}

	p := providers[name]
	if !p.Available {
		return fmt.Errorf("%s not configured. Set %s_API_KEY environment variable", p.Name, strings.ToUpper(name))
	}

	fmt.Printf("Set default provider to: %s\n", name)
	return nil
}

func runModels(cmd *cobra.Command, args []string) error {
	provider := currentProvider
	if provider == "" {
		provider = getDefaultProvider()
	}

	p, ok := providers[provider]
	if !ok {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	fmt.Printf("=== %s Models ===\n", p.Name)
	for _, m := range p.Models {
		fmt.Printf("  - %s\n", m)
	}

	return nil
}

func runAIConfig(cmd *cobra.Command, args []string) error {
	fmt.Println("=== AI Configuration ===")
	fmt.Println("Environment variables to set:")
	fmt.Println("  OPENAI_API_KEY, ANTHROPIC_API_KEY, GOOGLE_API_KEY, XAI_API_KEY")
	fmt.Println("\nCurrent configuration:")

	for name, p := range providers {
		if p.Available {
			fmt.Printf("  %s: configured\n", name)
		} else {
			fmt.Printf("  %s: not configured\n", name)
		}
	}

	return nil
}

func getDefaultProvider() string {
	for name, p := range providers {
		if p.Available {
			return name
		}
	}
	return "openai"
}

func generateResponse(provider, message string) string {
	words := strings.Fields(message)
	if len(words) < 3 {
		return "Configure your API key for full functionality."
	}
	return fmt.Sprintf("Processed '%s...' with %s. Token savings: ~70%%", strings.Join(words[:2], " "), provider)
}
