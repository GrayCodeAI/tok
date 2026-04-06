package dashboard

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/httpmw"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// AlertConfig holds alert threshold configuration
type AlertConfig struct {
	DailyTokenLimit     int64   `json:"daily_token_limit"`
	WeeklyTokenLimit    int64   `json:"weekly_token_limit"`
	UsageSpikeThreshold float64 `json:"usage_spike_threshold"` // multiplier for spike detection
	Enabled             bool    `json:"enabled"`
}

// Config holds dashboard configuration
type Config struct {
	Port             int         `json:"port"`
	Bind             string      `json:"bind"`
	UpdateInterval   int         `json:"update_interval"`
	Theme            string      `json:"theme"`
	Alerts           AlertConfig `json:"alerts"`
	EnableExport     bool        `json:"enable_export"`
	HistoryRetention int         `json:"history_retention"`
}

// DefaultConfig returns default dashboard configuration
var defaultConfig = Config{
	Port:             8080,
	Bind:             "localhost",
	UpdateInterval:   30000,
	Theme:            "dark",
	EnableExport:     true,
	HistoryRetention: 90,
	Alerts: AlertConfig{
		DailyTokenLimit:     1000000,
		WeeklyTokenLimit:    5000000,
		UsageSpikeThreshold: 2.0,
		Enabled:             true,
	},
}

var (
	Port   int
	Open   bool
	Bind   string
	APIKey string

	configMu sync.RWMutex
)

// Cmd returns the dashboard cobra command for registration.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Launch web dashboard for token savings visualization",
		Long: `Launch an interactive web dashboard to visualize token savings,
economics metrics, and usage trends over time.

The dashboard provides:
- Real-time token savings charts
- Daily/weekly/monthly breakdowns
- Command-level analytics
- Cost tracking with Claude API rates
- LLM usage integration via ccusage`,
		RunE: runDashboard,
	}

	cmd.Flags().IntVarP(&Port, "port", "p", 8080, "Port to run dashboard on")
	cmd.Flags().BoolVarP(&Open, "open", "O", false, "Open browser automatically")
	cmd.Flags().StringVar(&Bind, "bind", "localhost", "Address to bind server to (e.g., 0.0.0.0 for all interfaces)")
	cmd.Flags().StringVar(&APIKey, "api-key", "", "API key for authentication (empty = no auth)")

	return cmd
}

func runDashboard(cmd *cobra.Command, args []string) error {
	if Port < 1 || Port > 65535 {
		return fmt.Errorf("invalid --port %d: must be between 1 and 65535", Port)
	}

	dbPath := tracking.DatabasePath()
	tracker, err := tracking.NewTracker(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer tracker.Close()

	mux := http.NewServeMux()

	// API handlers
	mux.Handle("/api/stats", corsMiddleware(statsHandler(tracker)))
	mux.Handle("/api/context-reads", corsMiddleware(contextReadsHandler(tracker)))
	mux.Handle("/api/context-read-summary", corsMiddleware(contextReadSummaryHandler(tracker)))
	mux.Handle("/api/context-read-comparison", corsMiddleware(contextReadComparisonHandler(tracker)))
	mux.Handle("/api/context-read-quality", corsMiddleware(contextReadQualityHandler(tracker)))
	mux.Handle("/api/context-read-trend", corsMiddleware(contextReadTrendHandler(tracker)))
	mux.Handle("/api/context-read-top-files", corsMiddleware(contextReadTopFilesHandler(tracker)))
	mux.Handle("/api/context-read-projects", corsMiddleware(contextReadProjectsHandler(tracker)))
	mux.Handle("/api/daily", corsMiddleware(dailyHandler(tracker)))
	mux.Handle("/api/weekly", corsMiddleware(weeklyHandler(tracker)))
	mux.Handle("/api/monthly", corsMiddleware(monthlyHandler(tracker)))
	mux.Handle("/api/commands", corsMiddleware(commandsHandler(tracker)))
	mux.Handle("/api/recent", corsMiddleware(recentHandler(tracker)))
	mux.Handle("/api/economics", corsMiddleware(economicsHandler(tracker)))
	mux.Handle("/api/performance", corsMiddleware(performanceHandler(tracker)))
	mux.Handle("/api/failures", corsMiddleware(failuresHandler(tracker)))
	mux.Handle("/api/top-commands", corsMiddleware(topCommandsHandler(tracker)))
	mux.Handle("/api/hourly", corsMiddleware(hourlyHandler(tracker)))
	mux.Handle("/api/export/csv", corsMiddleware(exportCSVHandler(tracker)))
	// New endpoints for enhanced dashboard
	mux.Handle("/api/llm-status", corsMiddleware(llmStatusHandler(tracker)))
	mux.Handle("/api/daily-breakdown", corsMiddleware(dailyBreakdownHandler(tracker)))
	mux.Handle("/api/project-stats", corsMiddleware(projectStatsHandler(tracker)))
	mux.Handle("/api/session-stats", corsMiddleware(sessionStatsHandler(tracker)))
	mux.Handle("/api/savings-trend", corsMiddleware(savingsTrendHandler(tracker)))
	// New enhanced endpoints
	mux.Handle("/api/alerts", corsMiddleware(alertsHandler(tracker)))
	mux.Handle("/api/export/json", corsMiddleware(exportJSONHandler(tracker)))
	mux.Handle("/api/model-breakdown", corsMiddleware(modelBreakdownHandler(tracker)))
	mux.Handle("/api/config", corsMiddleware(configHandler(tracker)))
	mux.Handle("/api/report", corsMiddleware(reportHandler(tracker)))
	mux.Handle("/api/cache-metrics", corsMiddleware(cacheMetricsHandler(tracker)))
	// Chart and projection endpoints
	mux.Handle("/api/projection", corsMiddleware(costProjectionHandler(tracker)))
	mux.Handle("/api/contribution", corsMiddleware(contributionGraphHandler(tracker)))
	mux.Handle("/api/contribution-3d", corsMiddleware(contributionGraph3DHandler(tracker)))
	mux.HandleFunc("/logo", logoHandler)
	mux.HandleFunc("/", dashboardIndexHandler)

	rl := httpmw.NewDefault()

	var handler http.Handler = mux
	if APIKey != "" {
		handler = authMiddleware(APIKey, handler)
	}
	handler = rl.Middleware(handler)

	addr := fmt.Sprintf("%s:%d", Bind, Port)

	fmt.Printf("🌐 TokMan Dashboard running at http://%s\n", addr)
	fmt.Println("Press Ctrl+C to stop")

	if Open {
		fmt.Println("Opening browser...")
		// Could use browser.OpenURL here
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		if serveErr := srv.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-cmd.Context().Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			origin := r.Header.Get("Origin")
			if origin != "" && (strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1")) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health and static endpoints
		if r.URL.Path == "/" || r.URL.Path == "/logo" || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// inlineLogo is a minimal TokMan logo served from memory to avoid
// path-traversal risk from http.ServeFile with a relative path.
const inlineLogo = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" width="64" height="64">
  <rect width="64" height="64" rx="12" fill="#1a1a2e"/>
  <text x="32" y="42" text-anchor="middle" font-family="monospace" font-size="28" font-weight="bold" fill="#00d4ff">T</text>
</svg>`

func logoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	fmt.Fprint(w, inlineLogo)
}

func dashboardIndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, dashboardHTML)
}
