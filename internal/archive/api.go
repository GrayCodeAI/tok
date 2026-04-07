package archive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// APIServer provides REST API endpoints for the archive system
type APIServer struct {
	manager *ArchiveManager
	server  *http.Server
	addr    string
}

// NewAPIServer creates a new API server
func NewAPIServer(manager *ArchiveManager, addr string) *APIServer {
	if addr == "" {
		addr = ":8081"
	}

	api := &APIServer{
		manager: manager,
		addr:    addr,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/archives", api.handleArchives)
	mux.HandleFunc("/api/v1/archives/", api.handleArchive)
	mux.HandleFunc("/api/v1/archives/search", api.handleSearch)
	mux.HandleFunc("/api/v1/archives/stats", api.handleStats)
	mux.HandleFunc("/api/v1/archives/export", api.handleExport)
	mux.HandleFunc("/api/v1/archives/import", api.handleImport)
	mux.HandleFunc("/health", api.handleHealth)

	api.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return api
}

// Start starts the API server
func (api *APIServer) Start() error {
	slog.Info("starting archive API server", "addr", api.addr)
	return api.server.ListenAndServe()
}

// Stop gracefully stops the API server
func (api *APIServer) Stop(ctx context.Context) error {
	return api.server.Shutdown(ctx)
}

// handleArchives handles GET /api/v1/archives (list) and POST /api/v1/archives (create)
func (api *APIServer) handleArchives(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.listArchives(w, r)
	case http.MethodPost:
		api.createArchive(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handleArchive handles GET/DELETE /api/v1/archives/{hash}
func (api *APIServer) handleArchive(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/archives/")
	if path == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hash := strings.Split(path, "/")[0]

	switch r.Method {
	case http.MethodGet:
		api.getArchive(w, r, hash)
	case http.MethodDelete:
		api.deleteArchive(w, r, hash)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (api *APIServer) listArchives(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	opts := ArchiveListOptions{
		Limit:     100,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// Parse query params
	if limit := r.URL.Query().Get("limit"); limit != "" {
		opts.Limit, _ = strconv.Atoi(limit)
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		opts.Offset, _ = strconv.Atoi(offset)
	}
	if category := r.URL.Query().Get("category"); category != "" {
		opts.Category = ArchiveCategory(category)
	}
	if agent := r.URL.Query().Get("agent"); agent != "" {
		opts.Agent = agent
	}

	result, err := api.manager.List(ctx, opts)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":    result.Total,
		"has_more": result.HasMore,
		"entries":  result.Entries,
	})
}

func (api *APIServer) createArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Content     string   `json:"content"`
		Command     string   `json:"command"`
		Category    string   `json:"category"`
		Agent       string   `json:"agent"`
		Tags        []string `json:"tags"`
		ExpireHours int      `json:"expire_hours"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	entry := NewArchiveEntry([]byte(req.Content), req.Command)
	if req.Category != "" {
		entry.WithCategory(ArchiveCategory(req.Category))
	}
	if req.Agent != "" {
		entry.WithAgent(req.Agent)
	}
	if len(req.Tags) > 0 {
		entry.WithTags(req.Tags...)
	}
	if req.ExpireHours > 0 {
		entry.WithExpiration(time.Duration(req.ExpireHours) * time.Hour)
	}

	hash, err := api.manager.Archive(ctx, entry)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusCreated, map[string]string{
		"hash": hash,
	})
}

func (api *APIServer) getArchive(w http.ResponseWriter, r *http.Request, hash string) {
	ctx := r.Context()

	if !IsValidHash(hash) {
		api.writeError(w, http.StatusBadRequest, "invalid hash format")
		return
	}

	entry, err := api.manager.Retrieve(ctx, hash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			api.writeError(w, http.StatusNotFound, "archive not found")
			return
		}
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Check if raw content requested
	if r.URL.Query().Get("raw") == "true" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write(entry.OriginalContent)
		return
	}

	api.writeJSON(w, http.StatusOK, entry)
}

func (api *APIServer) deleteArchive(w http.ResponseWriter, r *http.Request, hash string) {
	ctx := r.Context()

	if !IsValidHash(hash) {
		api.writeError(w, http.StatusBadRequest, "invalid hash format")
		return
	}

	if err := api.manager.Delete(ctx, hash); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *APIServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	query := r.URL.Query().Get("q")
	if query == "" {
		api.writeError(w, http.StatusBadRequest, "query parameter 'q' required")
		return
	}

	opts := ArchiveListOptions{
		Query:  query,
		Limit:  50,
		SortBy: "created_at",
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		opts.Limit, _ = strconv.Atoi(limit)
	}

	result, err := api.manager.List(ctx, opts)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":   query,
		"total":   result.Total,
		"entries": result.Entries,
	})
}

func (api *APIServer) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	stats, err := api.manager.Stats(ctx)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, stats)
}

func (api *APIServer) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	var req struct {
		Hashes      []string `json:"hashes"`
		Format      string   `json:"format"`
		Compression bool     `json:"compression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	exporter := NewExporter(api.manager)
	opts := ExportOptions{
		Format:      ExportFormat(req.Format),
		Compression: req.Compression,
		IncludeRaw:  true,
	}

	var data []byte
	var filename string
	var err error

	if len(req.Hashes) == 0 {
		data, filename, err = exporter.ExportAll(ctx, opts, DefaultListOptions())
	} else if len(req.Hashes) == 1 {
		data, filename, err = exporter.Export(ctx, req.Hashes[0], opts)
	} else {
		data, filename, err = exporter.ExportMultiple(ctx, req.Hashes, opts)
	}

	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write(data)
}

func (api *APIServer) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Parse multipart form
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100MB max
		api.writeError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		api.writeError(w, http.StatusBadRequest, "file required")
		return
	}
	defer file.Close()

	// Read file to temp
	tempFile := fmt.Sprintf("/tmp/archive_import_%d", time.Now().Unix())
	f, err := os.Create(tempFile)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "failed to create temp file")
		return
	}
	defer os.Remove(tempFile)
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		api.writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	f.Close()

	importer := NewImporter(api.manager)
	result, err := importer.ImportFromFile(ctx, tempFile)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, result)
}

func (api *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	api.writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func (api *APIServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (api *APIServer) writeError(w http.ResponseWriter, status int, message string) {
	api.writeJSON(w, status, map[string]string{
		"error": message,
	})
}
