package web

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/proxy"
)

var (
	proxyListenAddr string
	proxyTargetURL  string
	proxyConfigFile string
	proxyDaemonize  bool
)

var proxyStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTP proxy server",
	Long: `Start the TokMan HTTP proxy server that intercepts LLM API calls
and compresses request messages before forwarding them.

Supports OpenAI, Anthropic, and Gemini API formats.

Examples:
  tokman http-proxy start
  tokman http-proxy start --listen :8080 --target https://api.openai.com
  tokman http-proxy start --config ~/.config/tokman/proxy.toml`,
	RunE: runProxyStart,
}

var proxyStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the HTTP proxy server",
	RunE:  runProxyStop,
}

var proxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show HTTP proxy server status",
	RunE:  runProxyStatus,
}

var httpProxyCmd = &cobra.Command{
	Use:   "http-proxy",
	Short: "HTTP proxy for LLM API compression",
	Long: `HTTP proxy that intercepts LLM API calls and compresses messages.
Works with OpenAI, Anthropic, and Gemini API formats.`,
}

func init() {
	registry.Add(func() { registry.Register(httpProxyCmd) })

	httpProxyCmd.AddCommand(proxyStartCmd)
	httpProxyCmd.AddCommand(proxyStopCmd)
	httpProxyCmd.AddCommand(proxyStatusCmd)

	proxyStartCmd.Flags().StringVarP(&proxyListenAddr, "listen", "l", ":8080", "Listen address")
	proxyStartCmd.Flags().StringVarP(&proxyTargetURL, "target", "t", "", "Target API URL")
	proxyStartCmd.Flags().StringVar(&proxyConfigFile, "config", "", "Config file path")
	proxyStartCmd.Flags().BoolVarP(&proxyDaemonize, "daemon", "d", false, "Run as daemon")
}

// ProxyConfig holds proxy server configuration.
type ProxyConfig struct {
	ListenAddr   string            `toml:"listen_addr"`
	TargetURL    string            `toml:"target_url"`
	ModelAliases map[string]string `toml:"model_aliases"`
	TLSEnabled   bool              `toml:"tls_enabled"`
	TLSCertFile  string            `toml:"tls_cert_file"`
	TLSKeyFile   string            `toml:"tls_key_file"`
	Enabled      bool              `toml:"enabled"`
}

func loadProxyConfig(path string) (*ProxyConfig, error) {
	cfg := &ProxyConfig{
		ListenAddr:   ":8080",
		TargetURL:    "https://api.openai.com",
		ModelAliases: make(map[string]string),
		Enabled:      true,
	}

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func runProxyStart(cmd *cobra.Command, args []string) error {
	cfg, err := loadProxyConfig(proxyConfigFile)
	if err != nil {
		return err
	}

	if proxyListenAddr != ":8080" {
		cfg.ListenAddr = proxyListenAddr
	}
	if proxyTargetURL != "" {
		cfg.TargetURL = proxyTargetURL
	}

	if !cfg.Enabled {
		return fmt.Errorf("proxy is disabled in config")
	}

	p := proxy.NewProxy(cfg.ListenAddr, cfg.TargetURL)

	// Apply model aliases
	for from, to := range cfg.ModelAliases {
		p.SetModelAlias(from, to)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      p,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Printf("TokMan HTTP Proxy listening on %s\n", cfg.ListenAddr)
		fmt.Printf("Target: %s\n", cfg.TargetURL)
		fmt.Printf("Health: http://%s/health\n", cfg.ListenAddr)
		fmt.Printf("Metrics: http://%s/metrics\n", cfg.ListenAddr)
		fmt.Println()

		var err error
		if cfg.TLSEnabled {
			err = server.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down proxy...")
	server.Close()
	return nil
}

func runProxyStop(cmd *cobra.Command, args []string) error {
	// Send SIGTERM to proxy process
	pidFile := os.Getenv("HOME") + "/.local/share/tokman/proxy.pid"
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("proxy not running (no PID file)")
	}

	var pid int
	fmt.Sscanf(string(data), "%d", &pid)
	if pid == 0 {
		return fmt.Errorf("invalid PID in file")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop proxy: %w", err)
	}

	fmt.Println("Proxy stopped")
	return nil
}

func runProxyStatus(cmd *cobra.Command, args []string) error {
	// Check if proxy is running
	pidFile := os.Getenv("HOME") + "/.local/share/tokman/proxy.pid"
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Println("Proxy: not running")
		return nil
	}

	var pid int
	fmt.Sscanf(string(data), "%d", &pid)

	process, err := os.FindProcess(pid)
	if err != nil || process.Signal(syscall.Signal(0)) != nil {
		fmt.Println("Proxy: not running (stale PID)")
		return nil
	}

	fmt.Printf("Proxy: running (PID %d)\n", pid)

	// Try to get stats from health endpoint
	resp, err := http.Get("http://localhost:8080/health")
	if err == nil {
		defer resp.Body.Close()
		fmt.Printf("Health: %s\n", resp.Status)
	}

	return nil
}
