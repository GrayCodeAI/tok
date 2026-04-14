package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/filter"
)

var (
	serverPort    string
	workspacePath string
	enableWS      bool
)

func init() {
	registry.Add(func() {
		registry.Register(serverCmd)
	})
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start API server with workspace support",
	Long: `Start TokMan REST API server with workspace management.
		
Workspaces allow multiple project contexts with separate:
- Command history and tracking
- Custom filters and configurations
- Session snapshots

Examples:
  tokman server serve --port 8081
  tokman server serve --workspace ~/projects/myapp`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	RunE:  runServe,
}

type Workspace struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	DBPath    string    `json:"db_path"`
}

type WorkspaceManager struct {
	workspaces map[string]*Workspace
	current    string
	mu         sync.RWMutex
	dataDir    string
}

func init() {
	serverCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&serverPort, "port", "p", "8081", "Server port")
	serveCmd.Flags().StringVar(&workspacePath, "workspace", "", "Workspace directory")
	serveCmd.Flags().BoolVar(&enableWS, "workspace-enabled", true, "Enable workspace support")
}

func NewWorkspaceManager() *WorkspaceManager {
	dataDir := filepath.Join(os.Getenv("HOME"), ".config/tokman", "workspaces")
	os.MkdirAll(dataDir, 0755)

	wm := &WorkspaceManager{
		workspaces: make(map[string]*Workspace),
		dataDir:    dataDir,
	}

	wm.loadWorkspaces()
	return wm
}

func (wm *WorkspaceManager) loadWorkspaces() {
	entries, err := os.ReadDir(wm.dataDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			wm.workspaces[name] = &Workspace{
				Name:      name,
				Path:      filepath.Join(wm.dataDir, name),
				CreatedAt: time.Now(),
				DBPath:    filepath.Join(wm.dataDir, name, "tokman.db"),
			}
		}
	}
}

func (wm *WorkspaceManager) Create(name string) (*Workspace, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, exists := wm.workspaces[name]; exists {
		return nil, fmt.Errorf("workspace already exists: %s", name)
	}

	wsPath := filepath.Join(wm.dataDir, name)
	if err := os.MkdirAll(wsPath, 0755); err != nil {
		return nil, err
	}

	ws := &Workspace{
		Name:      name,
		Path:      wsPath,
		CreatedAt: time.Now(),
		DBPath:    filepath.Join(wsPath, "tokman.db"),
	}

	wm.workspaces[name] = ws
	wm.current = name

	return ws, nil
}

func (wm *WorkspaceManager) List() []*Workspace {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	result := make([]*Workspace, 0, len(wm.workspaces))
	for _, ws := range wm.workspaces {
		result = append(result, ws)
	}
	return result
}

func (wm *WorkspaceManager) Switch(name string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, exists := wm.workspaces[name]; !exists {
		return fmt.Errorf("workspace not found: %s", name)
	}

	wm.current = name
	return nil
}

func (wm *WorkspaceManager) Current() string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.current
}

var workspaceManager *WorkspaceManager

func runServe(cmd *cobra.Command, args []string) error {
	if enableWS {
		workspaceManager = NewWorkspaceManager()
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/compress", handleCompress)
	mux.HandleFunc("/api/v1/estimate", handleEstimate)
	mux.HandleFunc("/api/v1/track", handleTrack)
	mux.HandleFunc("/api/v1/stats", handleStats)
	mux.HandleFunc("/api/v1/health", handleHealth)
	mux.HandleFunc("/api/v1/workspaces", handleWorkspaces)
	mux.HandleFunc("/v1/docs", handleOpenAPI)

	fmt.Printf("TokMan API server starting on port %s\n", serverPort)
	if enableWS && workspaceManager != nil {
		current := workspaceManager.Current()
		if current == "" {
			current = "default"
		}
		fmt.Printf("Workspace: %s\n", current)
	}

	srv := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down server...")
	srv.Shutdown(context.Background())

	return nil
}

func handleCompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Input     string `json:"input"`
		Mode      string `json:"mode"`
		Workspace string `json:"workspace"`
	}

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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"output":       output,
		"tokens_saved": saved,
		"tokens_left":  filter.EstimateTokens(output),
		"workspace":    req.Workspace,
	})
}

func handleEstimate(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"tokens": filter.EstimateTokens(req.Text)})
}

func handleTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_commands":     0,
		"total_tokens_saved": 0,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ws := ""
	if workspaceManager != nil {
		ws = workspaceManager.Current()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"version":   "0.28.2",
		"workspace": ws,
	})
}

func handleWorkspaces(w http.ResponseWriter, r *http.Request) {
	if workspaceManager == nil {
		http.Error(w, "workspace support not enabled", http.StatusNotImplemented)
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaceManager.List())
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ws, err := workspaceManager.Create(req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ws)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>TokMan API</title>
  <style>body{font-family:system-ui;padding:20px;background:#1a1a2e;color:#eee}</style>
</head>
<body>
<h1>TokMan API</h1>
<h2>Workspaces</h2>
<pre>
GET  /api/v1/workspaces           - List workspaces
POST /api/v1/workspaces           - Create workspace
</pre>
<h2>Compression</h2>
<pre>
POST /api/v1/compress - Compress text
POST /api/v1/estimate - Estimate tokens
POST /api/v1/track - Track command
GET  /api/v1/stats - Get statistics
GET  /api/v1/health - Health check
</pre>
</body>
</html>`))
}
