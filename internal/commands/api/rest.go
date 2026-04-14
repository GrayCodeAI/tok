package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/filter"
)

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start REST API server",
		Long:  `Start TokMan REST API server for programmatic access.`,
	}

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		RunE:  runServe,
	}

	httpPort  string
	certFile  string
	keyFile   string
	authToken string
)

func init() {
	serverCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&httpPort, "port", "p", "8081", "REST API port")
	serveCmd.Flags().StringVar(&certFile, "cert", "", "TLS certificate file")
	serveCmd.Flags().StringVar(&keyFile, "key", "", "TLS key file")
	serveCmd.Flags().StringVar(&authToken, "token", "", "API authentication token")

	registry.Add(func() {
		registry.Register(serverCmd)
	})
}

type APIServer struct {
	httpServer *http.Server
	wg         sync.WaitGroup
}

func runServe(cmd *cobra.Command, args []string) error {
	server := &APIServer{}

	server.wg.Add(1)
	go func() {
		defer server.wg.Done()
		server.startHTTPServer(httpPort)
	}()

	fmt.Printf("TokMan REST API server started on http://localhost:%s\n", httpPort)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /api/v1/compress - Compress text")
	fmt.Println("  POST /api/v1/estimate - Estimate tokens")
	fmt.Println("  POST /api/v1/track    - Track command")
	fmt.Println("  GET  /api/v1/stats    - Get statistics")
	fmt.Println("  GET  /api/v1/health   - Health check")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down server...")
	server.shutdown()
	server.wg.Wait()

	return nil
}

func (s *APIServer) startHTTPServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/compress", s.handleCompress)
	mux.HandleFunc("/api/v1/estimate", s.handleEstimate)
	mux.HandleFunc("/api/v1/track", s.handleTrack)
	mux.HandleFunc("/api/v1/stats", s.handleStats)
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/filter-test", s.handleFilterTest)

	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if certFile != "" && keyFile != "" {
		s.httpServer.ListenAndServeTLS(certFile, keyFile)
	} else {
		s.httpServer.ListenAndServe()
	}
}

func (s *APIServer) shutdown() {
	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
	}
}

type CompressRequest struct {
	Input  string `json:"input"`
	Mode   string `json:"mode"`
	Budget int    `json:"budget"`
	Preset string `json:"preset"`
}

type CompressResponse struct {
	Output       string  `json:"output"`
	TokensSaved  int     `json:"tokens_saved"`
	TokensLeft   int     `json:"tokens_left"`
	QualityScore float64 `json:"quality_score,omitempty"`
}

func (s *APIServer) handleCompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	if authToken != "" && r.Header.Get("Authorization") != "Bearer "+authToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mode := filter.ModeMinimal
	if req.Mode == "aggressive" {
		mode = filter.ModeAggressive
	}

	engine := filter.NewEngine(mode)
	output, saved := engine.Process(req.Input)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CompressResponse{
		Output:      output,
		TokensSaved: saved,
		TokensLeft:  filter.EstimateTokens(output),
	})
}

func (s *APIServer) handleEstimate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokens := filter.EstimateTokens(req.Text)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"tokens": tokens})
}

func (s *APIServer) handleTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *APIServer) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_commands":     0,
		"total_tokens_saved": 0,
	})
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": "0.28.2",
	})
}

func (s *APIServer) handleFilterTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Input  string `json:"input"`
		Output string `json:"output"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	originalTokens := filter.EstimateTokens(req.Input)
	filteredTokens := filter.EstimateTokens(req.Output)
	savings := float64(originalTokens-filteredTokens) / float64(originalTokens) * 100

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"original_tokens":   originalTokens,
		"filtered_tokens":   filteredTokens,
		"token_savings":     savings,
		"compression_ratio": float64(originalTokens) / float64(filteredTokens),
	})
}
